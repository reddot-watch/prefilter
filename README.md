# Security Event Prefilter

**A Purposefully Simple First-Line Filter for High-Volume Security Monitoring**

*(A deliberately "dumb" filter for smart systems - because sometimes you need a hammer, not a neural net)*

## The Prefilter Approach

Global security event monitoring requires effective first-pass filtering to manage high event volumes efficiently. This Go package implements a keyword-based filtering system with a strategically designed keyword list and matching approach. This would ensure near-certain capture of headline-worthy events while maintaining a false-positive-friendly approach.

## Why This Exists in the GPT Era

In an age of LLMs and neural networks, this tool embraces intentional simplicity for critical scenarios where:

- **Speed is survival**: Process thousands events/second on a potato-grade server
- **Resource constraints dominate**: Runs without GPU
- **False negatives are unacceptable**: Over-match rather than risk missing events
- **Explainability matters**: No black boxes - matches are directly traceable to keyword lists

This is not an AI system. This is **a simple sieve for security events** - designed to catch every possible grain of relevant information while letting obvious non-threats quickly pass through.

```ascii
Raw Event Firehose → [This Filter] → Reduced Stream → [ML/NLP/LLM Systems] → Analyst Review
                    (Fast 1st Pass)   (5-20% volume)    (Deep Filtering)
```

## Key Design Philosophy

- 🚦 **First-Stage Triage** - Reduce event volume before expensive processing
- 🚫 **No Machine Learning** - String matching only
- ⚡  **Microsecond Decisions** - Regex-free scanning
- 📜 **Explicit Rules** - No hidden model drift
- 🌍 **Language Agnostic** (But English-Focused*)
- 🔧 **Embracing False Positives** - By design

**Not a Replacement For:**
- Deep Learning Systems
- Semantic Analysis
- Large Language Models
- Human Judgment

**Optimized For:**
- Throughput
- Transparency
- Operational Simplicity

*While the architecture supports multiple languages, our current keyword list prioritizes English. Contributions for other languages are welcome via pull request.

## Features

- **Broad-Spectrum Keywords**: 2,000+ English terms covering physical security scenarios (direct terms, variants, indicators)
- **Basic Linguistic Coverage**: Includes common plurals & verb forms (`attack/attacks/attacked`)
- **Phrase Tolerance**: Matches "suspicious package" even in "suspicious looking package" or "suspicious unattended package"
- **Typo Resilience**: Handles `terorist` → `terrorist`, `hijaking` → `hijacking` using optimized Levenshtein distance
- **Multi-Language Ready** (Though English-Optimized)

## Installation

```bash
go get github.com/reddot-watch/prefilter
```

## Usage

```go
// Initialize with English keywords
filter, err := prefilter.NewFilter("en", prefilter.DefaultOptions())

// Sample news headline
text := "At least 11 dead in indonesian boat accidents"

if filter.Match(text) {
    // Pass to next stage for further processing
}
```

## Multi-Language Note

While the architecture supports multiple languages, our current keyword list prioritizes English. Contributions for other languages are welcome via pull request.

## License

This project is licensed under Apache 2.0 - See [LICENSE](LICENSE)

**Note:** The keyword lists may be subject to different licensing terms than the code itself. Please refer to the header comments in each keyword file (e.g., `keywords/en.txt`) for their specific license information.
