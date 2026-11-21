---
name: wine-registration
description: Register a wine and a physical bottle in CellarApp using wine-label information, without sending images through MCP, only after matching the cellar and receiving explicit confirmation.
---

# Wine registration for CellarApp

Use this Skill only with the CellarApp MCP server tools `search_wines`,
`register_bottle`, and `register_wine_and_bottle`.

## Extract and verify

If an image is available in the conversation, inspect it locally and record
producer, cuvee/name, vintage, country, region, appellation, and grape varieties
when legible. Never send, encode, or attach image bytes to an MCP tool call.
Mark each value as one of:

- **Visible on image**: directly readable from the user's image.
- **Verified online**: confirmed on an identified source; give the user the
  source name and URL.
- **Inferred**: plausible but not directly verified; never silently promote it
  to a fact.
- **Unknown**: not legible or not established. Keep vintage unknown rather
  than guessing.

When using web search, prefer the producer's official site, the relevant wine
region/appellation authority, or a reputable importer. Cross-check identity
and vintage; a search result snippet alone is not confirmation. Do not invent
master-data IDs: get IDs only from CellarApp's `list_countries`,
`list_regions`, `list_wine_types`, or `list_appellations` tools. Ask the user
when the required country or wine type cannot be identified confidently.

## Match against the cellar first

1. Call `search_wines` before any write. Search using producer and wine name;
   include vintage and country when known. If a query is too restrictive, try
   a second bounded query rather than treating an empty result as proof of
   absence.
2. Compare producer, wine name/cuvee, and vintage first. Country, region,
   appellation, and grapes are supporting evidence. Search results are
   partial-match candidates, not an identity verdict.
3. Different vintages are different candidates. A missing/unknown vintage,
   multiple plausible candidates, a producer/name similarity without enough
   evidence, or any material disagreement means **stop and ask the user**.
   Never write a new wine just because matching is ambiguous.
4. If there is exactly one well-supported existing wine, propose adding only a
   bottle to that `wine_id`. Do not edit the existing wine.
5. If the cellar search finds no plausible candidate and required fields are
   established, propose a new wine plus its first bottle. The label image will
   not be sent to or stored by the MCP registration tool.

## Approval is mandatory

Before calling either write tool, show the user:

- whether this will add a bottle to an existing wine (include wine ID and
  matching rationale) or create a new wine;
- all wine values with their evidence labels, including an explicit unknown
  vintage when applicable;
- bottle state, row, column, and note;
- for a new wine, state explicitly that no image will be sent or stored.

Ask for an explicit affirmative confirmation of those exact details. A
question, silence, uncertainty, or an ambiguous reply is not approval. If any
detail changes, present the revised details and obtain approval again. Never
call a write tool before approval. The server cannot authenticate a
conversation-level approval; this Skill and the tool descriptions are
operational safeguards, not a server-side authorization token.

## Image handling

The image may be inspected in the conversation to identify the wine, but it
must remain in the client and must never be sent to MCP, including as
base64-encoded text or any other image argument. The registration tools do not
accept or store label images; a newly registered wine will have no image URL
from this workflow. If the image cannot be inspected and required wine details
cannot otherwise be established, ask the user for the missing information.

## Write once and recover safely

- After approval, call exactly one relevant write tool once. Use
  `register_bottle` for a matched existing wine; use
  `register_wine_and_bottle` only for a clear new wine. Never call both for the
  same registration.
- Provide only known values and valid IDs returned by CellarApp. Do not fill
  unknown facts with guesses.
- An optional `idempotency_key` makes an identical request replay-safe. If the
  client provides a stable key, reuse it only for the identical approved
  request; do not reuse it for changed details. Keys are retained indefinitely.
- On timeout, disconnect, or an unknown outcome, do not blindly retry. Search
  the cellar first, compare possible results, and ask the user if a duplicate
  may have been created. A failed response is not proof that no write occurred.
- Read the success response, report the wine and bottle IDs, shelf position,
  and make no claim that an image was uploaded or associated.

Never update or delete existing wine records or bottles during this workflow.

## Hold / ask scenarios

Do not call either write tool in any of these cases:

| Situation | Required action |
| --- | --- |
| Same producer and cuvee but candidate vintage differs | Show candidates and ask; never treat it as the same wine |
| Producer or wine name is only approximately similar | Ask; no automatic match or new-wine fallback |
| More than one plausible cellar candidate | Present the differences and ask |
| Required country, wine type, producer/name identity, or bottle position is unknown | Ask for the missing information |
| Vintage is not legible and candidate identity depends on vintage | Keep vintage unknown and ask |
| Required wine details cannot be established from the available image or user-provided information | Ask for the missing information |
| User has not explicitly approved the exact final values | Do not call a write tool |
