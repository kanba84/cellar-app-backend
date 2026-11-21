package mcpserver

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type testReadService struct{ ReadOnlyService }
type testWriteService struct{ WriteService }

func TestRegistrationToolsPublishedWithSafeSchemasAndAnnotations(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	registerTools(server, testReadService{}, testWriteService{})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil).Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	tools, err := clientSession.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := clientSession.Close(); err != nil {
		t.Fatal(err)
	}
	if err := serverSession.Wait(); err != nil {
		t.Fatal(err)
	}

	found := map[string]*mcp.Tool{}
	for _, tool := range tools.Tools {
		found[tool.Name] = tool
	}
	for _, name := range []string{"list_wines", "search_wines", "register_bottle", "register_wine_and_bottle"} {
		if found[name] == nil {
			t.Errorf("tool %q was not published", name)
		}
	}
	if !found["list_wines"].Annotations.ReadOnlyHint {
		t.Error("list_wines must be marked read-only")
	}
	for _, name := range []string{"register_bottle", "register_wine_and_bottle"} {
		if found[name].Annotations.ReadOnlyHint {
			t.Errorf("%s must not be marked read-only", name)
		}
		if found[name].Annotations.DestructiveHint == nil || *found[name].Annotations.DestructiveHint {
			t.Errorf("%s should be declared additive, not destructive", name)
		}
		schema := found[name].InputSchema.(map[string]any)
		if schema["additionalProperties"] != false {
			t.Errorf("%s schema must reject additional properties", name)
		}
	}
	searchSchema := found["search_wines"].InputSchema.(map[string]any)
	if searchSchema["additionalProperties"] != false {
		t.Error("search_wines schema must reject additional properties")
	}
	newWineSchema := found["register_wine_and_bottle"].InputSchema.(map[string]any)
	properties := newWineSchema["properties"].(map[string]any)
	wineSchema := properties["wine"].(map[string]any)
	if wineSchema["additionalProperties"] != false {
		t.Error("nested wine schema must reject additional properties")
	}
	if _, ok := wineSchema["required"]; !ok {
		t.Error("nested wine schema must declare required fields")
	}
	for _, imageField := range []string{"label_image_base64", "label_image_mime_type"} {
		if _, ok := properties[imageField]; ok {
			t.Errorf("registration schema must not accept image field %q", imageField)
		}
	}
	if !strings.Contains(found["register_wine_and_bottle"].Description, "without sending or storing a label image") {
		t.Error("registration tool description must state that images are not sent or stored")
	}
}

func TestMCPInputValidationBoundaries(t *testing.T) {
	if err := validateWineSearch(searchWinesInput{}); err == nil {
		t.Fatal("empty wine search must be rejected")
	}
	if err := validateWineSearch(searchWinesInput{Producer: "Domaine", Vintage: intPointer(2020), Limit: 20}); err != nil {
		t.Fatalf("bounded partial-match search should pass: %v", err)
	}
	if err := validateWineSearch(searchWinesInput{Name: "Wine", Limit: 21}); err == nil {
		t.Fatal("search limit above 20 must be rejected")
	}
	if err := validateRegisterBottle(registerBottleInput{WineID: 1, RowNumber: 10, ColumnNumber: 1}); err != nil {
		t.Fatalf("valid shelf boundary should pass: %v", err)
	}
	if err := validateRegisterBottle(registerBottleInput{WineID: 1, RowNumber: 11, ColumnNumber: 1}); err == nil {
		t.Fatal("row above cellar limit must be rejected")
	}
	if err := validateRegisterWineAndBottle(registerWineAndBottleInput{
		Wine:   registrationWineInput{Name: "Wine", CountryID: 1, WineTypeID: 1, GrapeNames: []string{"Merlot", "merlot"}},
		Bottle: registrationBottleInput{RowNumber: 1, ColumnNumber: 1},
	}); err == nil {
		t.Fatal("duplicate grape names must be rejected")
	}
}

func TestMCPHTTPBodySizeLimit(t *testing.T) {
	handler := NewHTTPHandler(testReadService{}, testWriteService{})
	body := bytes.NewReader(make([]byte, maxMCPRequestBodyBytes+1))
	request := httptest.NewRequest(http.MethodPost, "/mcp", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized MCP request status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestSafeWriteErrorHidesInternalDetails(t *testing.T) {
	err := safeWriteError(errors.New("postgres://secret.internal/db table cellar_bottles"))
	if strings.Contains(err.Error(), "secret.internal") || strings.Contains(err.Error(), "cellar_bottles") {
		t.Fatalf("internal detail leaked in tool error: %v", err)
	}
}

func TestSkillDocumentsHoldScenarios(t *testing.T) {
	content, err := os.ReadFile("../wine-registration/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Different vintages are different candidates",
		"More than one plausible cellar candidate",
		"conversation-level approval",
		"do not blindly retry",
		"must never be sent to MCP",
	} {
		if !bytes.Contains([]byte(strings.Join(strings.Fields(string(content)), " ")), []byte(expected)) {
			t.Errorf("Skill missing required hold instruction %q", expected)
		}
	}
}

func intPointer(value int) *int { return &value }
