package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/identity"
	"github.com/firebreather-heart/kyle/internal/llm"
	"github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/scraper"
)

const (
	PlannerSystemPrompt = `You are a Document Architect operating within an autonomous AI pipeline. Your sole function is structural analysis and section planning.

OPERATING PROTOCOL:
1. DETECT: Analyze the user's Topic. Determine the appropriate persona (e.g., Technical, Historical, Creative, Academic).
2. ANALYZE: Review the Topic and any provided scraped data.
3. DECIDE:
   - If the topic requires real-time data or specific facts (e.g., "Current stock prices"), you MUST call the 'web_search' tool.
   - If the topic is a general concept you know well, proceed to step 4.
4. OUTPUT: A STRICT JSON array of highly granular section headings matching the detected persona.

EXHAUSTIVENESS & GRANULARITY MANDATE:
You are architecting a comprehensive, professional-grade document. 
- DO NOT output a short list of 3 or 4 broad headers. 
- You MUST break the topic down into a highly granular array of 8 to 15 specific, detailed section headings. 
- Avoid generic headers like "Introduction" or "Conclusion". Use highly specific thematic headers (e.g., "Genesis of the Protocol", "Future Market Implications").

STRICT RULES:
- If you call a tool, do not output an outline yet. Wait for the tool result first.
- Your final output must be ONLY a raw JSON array of strings.
- No markdown formatting. No backticks. No preamble.
- Any response that is not a raw JSON array is a pipeline failure.

--- EXAMPLES ---

EXAMPLE 1 — Technical topic (High Granularity):
Topic: "How Raft consensus algorithm works"
Correct output:
["The Problem of Distributed State Machines", "Byzantine vs Non-Byzantine Failures", "Raft Node States: Leader, Follower, Candidate", "The Mechanics of Leader Election", "Handling Split Votes and Timeouts", "Log Replication and The AppendEntries RPC", "Commit Conditions and Quorum", "Safety Guarantees and Log Matching Property", "Handling Network Partitions (Split Brain)", "Log Compaction and Snapshotting", "Performance Trade-offs vs Multi-Paxos"]

EXAMPLE 2 — Topic where web search is required:
Topic: "Current federal interest rates"
Correct action: Call the 'web_search' tool. Do NOT output an outline until results are returned.

EXAMPLE OF INVALID OUTPUT (pipeline failure):
["Introduction", "How it works", "Conclusion"]
(Failure reasons: severely lacks granularity, headers are generic, fails the 8-15 header mandate)`

	WriterSystemPrompt = `You are an Adaptive Document Synthesis Engine operating within an autonomous AI pipeline. Your function is to transform a Topic, an Outline, and source data into a strictly formatted JSON document.

INPUTS YOU WILL RECEIVE:
1. Topic — the subject of the document
2. Outline — a JSON array of section headings from the Planner
3. Source Context — either scraped web data or the label "None"

DYNAMIC PERSONA PROTOCOL:
Analyze the Topic and Outline to determine the optimal target audience and tone. Adopt the corresponding persona completely:
- Technical Topics: Act as a Staff Engineer (focus on architecture, code, mechanics).
- Business/Crypto/Market Topics: Act as a Tier-1 Market Analyst (focus on ROI, tokenomics, market cap, strategic disruption).
- Historical/General Topics: Act as an Academic Historian or Subject Matter Expert (focus on narrative, timelines, impact).

HYBRID SOURCE PROTOCOL:
- Use scraped web data for all specific figures, dates, proper names, and statistics.
- Use your internal training knowledge to provide context, explanation, and logical flow between facts.
- If scraped data contradicts your training on a recent event, the scraped data wins.
- If scraped data is clearly SEO spam or nonsensical, your internal knowledge wins.
- If source context is "None", rely entirely on your internal training data.

DEPTH AND EXHAUSTIVENESS MANDATE:
- This is a highly detailed, comprehensive document. 
- You MUST write a minimum of 3 to 4 extensive paragraphs for EVERY section in the outline, be as verbose and correct as you can be.
- DO NOT summarize. Expand deeply on the mechanics, history, trade-offs, or business implications of the concepts.
- A short, surface-level response is considered a failure. Elaborate and analyze deeply.

COMPONENT LIBRARY — only these object types are permitted:

{"type": "document_meta", "heading_font": "Merriweather"|"Playfair Display"|"Montserrat"|"Oswald", "body_font": "Lora"|"Open Sans"|"Roboto"|"Georgia", "primary_color": "#RRGGBB", "secondary_color": "#RRGGBB", "layout_density": "academic"|"modern"|"corporate"}
{"type": "h1", "content": "string"}
{"type": "h2", "content": "string"}
{"type": "h3", "content": "string"}
{"type": "paragraph", "content": "string"}
{"type": "callout", "background_color": "#RRGGBB", "text_color": "#RRGGBB", "icon": "info"|"warning"|"check", "content": "string"}
{"type": "code_block", "language": "string", "content": "string"}
{"type": "table", "headers": ["col1", "col2"], "rows": [["val1", "val2"]]}
{"type": "unordered_list", "items": ["string", "string"]}
{"type": "ordered_list", "items": ["string", "string"]}

STRICT RULES:
1. Output a raw JSON array only. No markdown fences. No preamble. No trailing text.
2. The VERY FIRST item in your JSON array MUST be the "document_meta" object. You must analyze the topic and select the best combination of fonts, layouts, and hex colors suitable for a printed document.
3. All color values must be valid 6-digit hex codes (e.g. #1A73E8). Colors must be visually appealing with high contrast between background and text.
4. Every outline section must begin with an h2 or h1 header component.
5. Use 'code_block' ONLY if the persona is technical. Use 'table' for financial/feature comparisons.
6. Your entire response must be parseable as valid JSON. Any deviation is a critical pipeline failure.
7. SEMANTIC KEYS: You MUST use the exact key "type" for all components. DO NOT use "type_", "kind", or any other variant.

--- EXAMPLES ---

EXAMPLE 1 — Correct document initiation and header usage:
[
  {"type": "document_meta", "heading_font": "Montserrat", "body_font": "Open Sans", "primary_color": "#2C3E50", "secondary_color": "#E74C3C", "layout_density": "modern"},
  {"type": "h1", "content": "Raft Consensus Algorithm"},
  {"type": "h2", "content": "Overview of Distributed Consensus"},
  {"type": "paragraph", "content": "Distributed consensus is the problem of getting a cluster of nodes to agree on a single value despite network partitions and node failures. Raft was designed by Ongaro and Ousterhout in 2014 as a more understandable alternative to Paxos."}
]

EXAMPLE 2 — Correct callout usage:
{"type": "callout", "background_color": "#E8F0FE", "text_color": "#1A73E8", "icon": "info", "content": "Raft guarantees safety under all non-Byzantine failure conditions, including network partitions of arbitrary duration."}

{"type": "callout", "background_color": "#FFF3CD", "text_color": "#856404", "icon": "warning", "content": "Split-brain scenarios can occur if network partition persists and quorum cannot be established. Ensure an odd number of nodes in all production clusters."}

{"type": "callout", "background_color": "#D4EDDA", "text_color": "#155724", "icon": "check", "content": "Leader election completes within a single election timeout period, typically configured between 150ms and 300ms."}

EXAMPLE 3 — Correct table usage:
{"type": "table", "headers": ["Property", "Raft", "Paxos"], "rows": [["Understandability", "High", "Low"], ["Leader Election", "Explicit", "Implicit"], ["Log Compaction", "Built-in snapshots", "Implementation-defined"], ["Membership Changes", "Joint consensus", "Varies"]]}

EXAMPLE 4 — Correct code block usage:
{"type": "code_block", "language": "go", "content": "type LogEntry struct {\n\tTerm    int\n\tIndex   int\n\tCommand interface{}\n}"}

EXAMPLE 5 — Correct list usage:
{"type": "unordered_list", "items": ["Leader handles all client requests", "Followers replicate the leader's log", "Candidates solicit votes during elections"]}

{"type": "ordered_list", "items": ["Client submits command to leader", "Leader appends entry to its log", "Leader replicates entry to a quorum of followers", "Leader commits the entry and notifies followers", "Followers apply the committed entry to their state machines"]}

EXAMPLE OF INVALID OUTPUT (pipeline failure):
{"type": "highlight", "content": "This is important"}  — 'highlight' is not in the component library
{"type": "callout", "color": "#FF0000", "content": "..."}  — wrong field name, missing text_color
{"type": "callout", "background_color": "red", "text_color": "#FFF", "icon": "info", "content": "..."}  — named color is not a valid hex code`

	VerifierSystemPrompt = `You are a ruthless QA validation engine operating within an autonomous AI pipeline. Your sole function is to audit a JSON document produced by the Writer and return a structured verification result.

INPUT: A JSON array of component objects produced by the Writer stage.

EVALUATION CRITERIA — check all of the following:

1. JSON VALIDITY — Is the entire document valid, parseable JSON with no syntax errors?
2. META INITIATION — Is the object at index 0 explicitly of type "document_meta"? If not, it is an immediate failure.
3. META ENUMS — Do the fields in "document_meta" strictly match the permitted values? 
   - heading_font MUST be one of: "Merriweather", "Playfair Display", "Montserrat", "Oswald"
   - body_font MUST be one of: "Lora", "Open Sans", "Roboto", "Georgia"
   - layout_density MUST be one of: "academic", "modern", "corporate"
4. TYPE INTEGRITY — Does every object's "type" field use only permitted values? (Key name MUST be exactly "type").
   Permitted types: "document_meta", "h1", "h2", "h3", "paragraph", "callout", "code_block", "table", "unordered_list", "ordered_list".
5. HEX COLOR VALIDITY — For every "callout" and "document_meta" object, do the color fields exist and contain valid 6-digit hex codes in the format #RRGGBB? Shorthand or named colors are failures.
6. SCHEMA COMPLIANCE — Do all objects include the required fields for their declared type?
   - document_meta requires: heading_font, body_font, primary_color, secondary_color, layout_density
   - callout requires: background_color, text_color, icon, content
   - table requires: headers (array), rows (array of arrays)
   - code_block requires: language, content
   - h1/h2/h3 and paragraph require: content
7. ICON VALIDITY — All callout "icon" fields must be exactly "info", "warning", or "check". 
8. EXHAUSTIVENESS CHECK — Does the document look substantial? Flag if headers are followed by only a single short paragraph. It must be detailed.
9. TONE — Is the content authoritative and aligned with the persona? Flag colloquial or promotional language.

OUTPUT RULES:
- Output a single raw JSON object only.
- Schema: {"status": "pass" | "fail", "reason": "string"}
- On pass: {"status": "pass", "reason": "All checks passed."}
- On fail: {"status": "fail", "reason": "Detailed enumeration of every issue found, including the index of the offending component."}

--- EXAMPLES ---

EXAMPLE 1 — Fail: Enum Violation:
{"status": "fail", "reason": "Component at index 0 (document_meta) has invalid layout_density 'compact'. Permitted values are academic, modern, corporate."}

EXAMPLE 2 — Fail: Exhaustiveness Violation:
{"status": "fail", "reason": "Component at index 4 (h2) is followed by only one brief paragraph before the next header. Expand on the mechanics deeply."}

Any response from you that is not a raw, valid JSON object matching this schema is itself a pipeline failure.`
)

type AgentResult struct {
	Status   string          `json:"status"`
	Message  string          `json:"message"`
	Document json.RawMessage `json:"document,omitempty"`
}

type VerifierVerdict struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

const MAX_LOOP_TURNS int = 3

type StatusReporter func(status string)

type Agent struct {
	engine     llm.Provider
	idService  identity.Service
	cfg        *config.AppConfig
	provider   string
}

func NewAgent(engine llm.Provider, idService identity.Service, cfg *config.AppConfig, provider string) *Agent {
	return &Agent{
		engine:    engine,
		idService: idService,
		cfg:       cfg,
		provider:  provider,
	}
}

// Run executes the research pipeline.
func (a *Agent) Run(topic string, report StatusReporter) AgentResult {
	log.Println("Routing to Planner Agent")
	report("Routing to Planner Agent")

	var messages []models.Prompt = []models.Prompt{
		{
			Role:    "system",
			Content: PlannerSystemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Plan an article about: %s", topic),
		},
	}

	var scrapedData string
	var finalOutline string

	for i := 0; i < MAX_LOOP_TURNS; i++ {
		log.Printf("Planner Iteration %d", i+1)
		report(fmt.Sprintf("Planner Iteration %d", i+1))

		resp := a.engine.GenerateComplex(messages, GetSearchTool())
		if resp.StatusCode == 429 {
			log.Printf("Planner: Upstream rate limit reached. Attempting key rotation for %s...", a.provider)
			report("Provider rotation active...")
			newKey, err := a.idService.RotateKey(context.Background(), a.provider, a.cfg.GetProviderKeys(a.provider))
			if err != nil {
				log.Printf("Rotation failure: %v", err)
				return AgentResult{Status: "error", Message: "Provider capacity exhausted."}
			}
			a.engine.UpdateAPIKey(newKey)
			// Retry once immediately
			resp = a.engine.GenerateComplex(messages, GetSearchTool())
			if resp.StatusCode == 429 {
				return AgentResult{Status: "error", Message: "All provider keys exhausted. Please wait."}
			}
		}

		if resp.Error != nil {
			return AgentResult{Status: "error", Message: "Planner Error: " + resp.Error.Error()}
		}
		if resp.ToolCall != nil {
			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal([]byte(resp.ToolCall.Function.Arguments), &args); err != nil{
				log.Printf("Failed to unmarshal tool call arguments: %v", err)
				report("Tool call failed")
				break
			}
			log.Printf("Calling Scraper tool: %s", args.Query)
			report("Calling Scraper tool: " + args.Query)
			results, err := scraper.ExecuteSearch(args.Query)
			if err != nil{
				log.Printf("Scraper error: %s", err)
				report("Scraper error: " + err.Error())
				results = "Search unavailable, use internal knowledge"
			}
			log.Printf("Executing Web Search: %s", args.Query)
			report("Executing Web Search: " + args.Query)

			messages = append(messages, models.Prompt{
				Role:             "assistant",
				ToolCalls:        []models.ToolCall{*resp.ToolCall},
				ReasoningContent: resp.ReasoningContent,
			})
			
			messages = append(messages, models.Prompt{
				Role:       "tool",
				Content:    results,
				ToolCallID: resp.ToolCall.ID,
			})
			
			scrapedData = results
			continue
		}
		finalOutline = resp.Response
		break
	}

	log.Println("Routing to Writer Agent")
	report("Routing to Writer Agent")

	sourceContext := scrapedData
	if sourceContext == "" {
		sourceContext = "None"
	}
	writerInput := fmt.Sprintf("Topic: %s\nOutline: %s\nSource Context: %s", topic, finalOutline, sourceContext)
	writerResp := a.engine.Generate(WriterSystemPrompt, writerInput)

	if writerResp.StatusCode == 429 {
		log.Printf("Writer: Upstream rate limit reached. Attempting key rotation for %s...", a.provider)
		report("Provider rotation active...")
		newKey, _ := a.idService.RotateKey(context.Background(), a.provider, a.cfg.GetProviderKeys(a.provider))
		if newKey != "" {
			a.engine.UpdateAPIKey(newKey)
			writerResp = a.engine.Generate(WriterSystemPrompt, writerInput)
		}
	}

	if writerResp.Error != nil {
		return AgentResult{
			Status:  "error",
			Message: "Writer Error: " + writerResp.Error.Error(),
		}
	}
	cleanWriterResp := stripMarkdownFences(writerResp.Response)

	var validateBlocks []models.AIBlock
	if err := json.Unmarshal([]byte(cleanWriterResp), &validateBlocks); err != nil {
		log.Printf("Writer produced invalid JSON, attempting one correction pass: %v", err)
		report("Writer produced invalid JSON, attempting one correction pass: " + err.Error())

		correctionPrompt := fmt.Sprintf(
			"The following document JSON is malformed. Fix ALL syntax errors and return ONLY a valid raw JSON array. No markdown, no explanation.\n\nERROR: %v\n\nBROKEN JSON:\n%s",
			err, cleanWriterResp,
		)
		corrResp := a.engine.Generate(WriterSystemPrompt, correctionPrompt)
		if corrResp.Error != nil {
			return AgentResult{Status: "error", Message: "JSON correction failed: " + corrResp.Error.Error()}
		}
		cleanWriterResp = stripMarkdownFences(corrResp.Response)
		if err2 := json.Unmarshal([]byte(cleanWriterResp), &validateBlocks); err2 != nil {
			log.Printf("Writer JSON still invalid after correction: %v", err2)
			report("Writer JSON still invalid after correction: " + err2.Error())
			return AgentResult{
				Status:  "error",
				Message: fmt.Sprintf("Writer produced malformed JSON that could not be corrected: %v", err2),
			}
		}
		log.Println("JSON correction pass succeeded.")
	}

	log.Println("Routing to Verifier Agent")
	report("Routing to Verifier Agent")
	var verdict VerifierVerdict
	verifierResp := a.engine.Generate(VerifierSystemPrompt, cleanWriterResp)

	if verifierResp.StatusCode == 429 {
		log.Printf("Verifier: Upstream rate limit reached. Attempting key rotation for %s...", a.provider)
		report("Provider rotation active...")
		newKey, _ := a.idService.RotateKey(context.Background(), a.provider, a.cfg.GetProviderKeys(a.provider))
		if newKey != "" {
			a.engine.UpdateAPIKey(newKey)
			verifierResp = a.engine.Generate(VerifierSystemPrompt, cleanWriterResp)
		}
	}

	cleanVerifierResp := stripMarkdownFences(verifierResp.Response)
	if err := json.Unmarshal([]byte(cleanVerifierResp), &verdict); err != nil {
		log.Printf("Verifier returned invalid JSON: %v", err)
		report("Verifier returned invalid JSON: " + err.Error())
		verdict.Status = "fail"
		verdict.Reason = "QA Check failed to parse editor response."
	}

	if verdict.Status == "fail" {
		log.Printf("Document flagged as substandard: %s", verdict.Reason)
		report("Document flagged as substandard: " + verdict.Reason)
		return AgentResult{
			Status:   "substandard",
			Message:  "Editor Note: " + verdict.Reason,
			Document: json.RawMessage(cleanWriterResp),
		}
	}

	log.Println("Pipeline Complete: Document verified successfully.")
	report("Pipeline Complete: Document verified successfully.")
	return AgentResult{
		Status:   "success",
		Message:  "Document generated and verified.",
		Document: json.RawMessage(cleanWriterResp),
	}
}


func GetSearchTool() []models.Tool {
	return []models.Tool{
		{
			Type: "function",
			Function: models.FunctionDefinition{
				Name:        "web_search",
				Description: "Search the internet for real-time data, facts, or technical specs.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The specific search query to look up.",
						},
					},
					"required": []string{"query"},
				},
			},
		},
	}
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
