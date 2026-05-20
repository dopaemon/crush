package prompt

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/filepathext"
	"github.com/charmbracelet/crush/internal/home"
	"github.com/charmbracelet/crush/internal/shell"
	"github.com/charmbracelet/crush/internal/skills"
)

// Prompt represents a template-based prompt generator.
type Prompt struct {
	name       string
	template   string
	now        func() time.Time
	platform   string
	workingDir string
	toolNames  []string
	isSubAgent bool
	sections   *sectionCache
}

type PromptDat struct {
	Provider      string
	Model         string
	Config        config.Config
	WorkingDir    string
	IsGitRepo     bool
	Platform      string
	Date          string
	GitStatus     string
	ContextFiles  []ContextFile
	AvailSkillXML string
	ToolGuidance  string
	LanguageSection string
	OutputStyleSection string
	MemorySection string
	MCPInstructionsSection string
	DynamicBoundary string
}

type ContextFile struct {
	Path    string
	Content string
}

type Option func(*Prompt)

func WithTimeFunc(fn func() time.Time) Option {
	return func(p *Prompt) {
		p.now = fn
	}
}

func WithPlatform(platform string) Option {
	return func(p *Prompt) {
		p.platform = platform
	}
}

func WithWorkingDir(workingDir string) Option {
	return func(p *Prompt) {
		p.workingDir = workingDir
	}
}

func WithToolNames(toolNames []string) Option {
	return func(p *Prompt) {
		p.toolNames = append([]string(nil), toolNames...)
	}
}

func WithSubAgent(isSubAgent bool) Option {
	return func(p *Prompt) {
		p.isSubAgent = isSubAgent
	}
}

func NewPrompt(name, promptTemplate string, opts ...Option) (*Prompt, error) {
	p := &Prompt{
		name:     name,
		template: promptTemplate,
		now:      time.Now,
		sections: newSectionCache(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

// ClearSectionCache resets memoized prompt sections.
// Useful after external state changes that should invalidate cached sections.
func (p *Prompt) ClearSectionCache() {
	if p.sections != nil {
		p.sections.clear()
	}
}

func (p *Prompt) Build(ctx context.Context, provider, model string, store *config.ConfigStore) (string, error) {
	t, err := template.New(p.name).Parse(p.template)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var sb strings.Builder
	d, err := p.promptData(ctx, provider, model, store)
	if err != nil {
		return "", err
	}
	if err := t.Execute(&sb, d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return sb.String(), nil
}

func processFile(filePath string) *ContextFile {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return &ContextFile{
		Path:    filePath,
		Content: string(content),
	}
}

func processContextPath(p string, store *config.ConfigStore) []ContextFile {
	var contexts []ContextFile
	fullPath := filepathext.SmartJoin(store.WorkingDir(), p)
	info, err := os.Stat(fullPath)
	if err != nil {
		return contexts
	}
	if info.IsDir() {
		filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				if result := processFile(path); result != nil {
					contexts = append(contexts, *result)
				}
			}
			return nil
		})
	} else {
		result := processFile(fullPath)
		if result != nil {
			contexts = append(contexts, *result)
		}
	}
	return contexts
}

// expandPath expands ~ and environment variables in file paths
func expandPath(path string, store *config.ConfigStore) string {
	path = home.Long(path)
	// Handle environment variable expansion using the same pattern as config
	if strings.HasPrefix(path, "$") {
		if expanded, err := store.Resolver().ResolveValue(path); err == nil {
			path = expanded
		}
	}

	return path
}

func (p *Prompt) promptData(ctx context.Context, provider, model string, store *config.ConfigStore) (PromptDat, error) {
	workingDir := cmp.Or(p.workingDir, store.WorkingDir())
	platform := cmp.Or(p.platform, runtime.GOOS)

	files := map[string][]ContextFile{}

	cfg := store.Config()
	for _, pth := range cfg.Options.ContextPaths {
		expanded := expandPath(pth, store)
		pathKey := strings.ToLower(expanded)
		if _, ok := files[pathKey]; ok {
			continue
		}
		content := processContextPath(expanded, store)
		files[pathKey] = content
	}

	// Discover and load skills metadata.
	var availSkillXML string

	// Start with builtin skills.
	allSkills := skills.DiscoverBuiltin()
	builtinNames := make(map[string]bool, len(allSkills))
	for _, s := range allSkills {
		builtinNames[s.Name] = true
	}

	// Discover user skills from configured paths.
	if len(cfg.Options.SkillsPaths) > 0 {
		expandedPaths := make([]string, 0, len(cfg.Options.SkillsPaths))
		for _, pth := range cfg.Options.SkillsPaths {
			expandedPaths = append(expandedPaths, expandPath(pth, store))
		}
		for _, userSkill := range skills.Discover(expandedPaths) {
			if builtinNames[userSkill.Name] {
				slog.Warn("User skill overrides builtin skill", "name", userSkill.Name)
			}
			allSkills = append(allSkills, userSkill)
		}
	}

	// Deduplicate: user skills override builtins with the same name.
	allSkills = skills.Deduplicate(allSkills)

	// Filter out disabled skills.
	allSkills = skills.Filter(allSkills, cfg.Options.DisabledSkills)

	if len(allSkills) > 0 {
		availSkillXML = skills.ToPromptXML(allSkills)
	}

	isGit := isGitRepo(store.WorkingDir())
	data := PromptDat{
		Provider:      provider,
		Model:         model,
		Config:        *cfg,
		WorkingDir:    filepath.ToSlash(workingDir),
		IsGitRepo:     isGit,
		Platform:      platform,
		Date:          p.now().Format("1/2/2006"),
		AvailSkillXML: availSkillXML,
		ToolGuidance:  "",
		DynamicBoundary: SystemPromptDynamicBoundary,
	}
	sections := []promptSection{
		systemPromptSection("tool_guidance", func(_ PromptDat) string {
			return renderToolGuidance(p.toolNames, p.isSubAgent)
		}),
		systemPromptSection("language", func(d PromptDat) string {
			if strings.TrimSpace(d.Config.Options.Language) == "" {
				return ""
			}
			lang := strings.TrimSpace(d.Config.Options.Language)
			return "# Language\nAlways respond in " + lang + ". Use " + lang + " for all explanations and communication with the user. Keep code and identifiers unchanged."
		}),
		systemPromptSection("output_style", func(d PromptDat) string {
			style := strings.TrimSpace(d.Config.Options.OutputStylePrompt)
			if style == "" {
				return ""
			}
			return "# Output Style\n" + style
		}),
		systemPromptSection("memory", func(d PromptDat) string {
			if len(d.ContextFiles) == 0 {
				return ""
			}
			return "# Memory\nFollow instructions from injected memory/context files (`<memory>` blocks). Treat them as durable user/project preferences unless explicitly overridden by higher-priority instructions."
		}),
		uncachedSystemPromptSection("mcp_instructions", func(_ PromptDat) string {
			// Runtime MCP instructions are appended per-turn in agent.go.
			// Keep this section uncached so dynamic prompt architecture mirrors
			// Claude-style cache-break semantics for volatile server instructions.
			return ""
		}),
	}
	for _, s := range sections {
		if s.cacheBreak {
			if s.name == "mcp_instructions" {
				data.MCPInstructionsSection = s.compute(data)
			}
			continue
		}
		if v, ok := p.sections.get(s.name); ok {
			switch s.name {
			case "tool_guidance":
				data.ToolGuidance = v
			case "language":
				data.LanguageSection = v
			case "output_style":
				data.OutputStyleSection = v
			case "memory":
				data.MemorySection = v
			}
			continue
		}
		v := s.compute(data)
		p.sections.set(s.name, v)
		switch s.name {
		case "tool_guidance":
			data.ToolGuidance = v
		case "language":
			data.LanguageSection = v
		case "output_style":
			data.OutputStyleSection = v
		case "memory":
			data.MemorySection = v
		}
	}
	if isGit {
		var err error
		data.GitStatus, err = getGitStatus(ctx, store.WorkingDir())
		if err != nil {
			return PromptDat{}, err
		}
	}

	for _, contextFiles := range files {
		data.ContextFiles = append(data.ContextFiles, contextFiles...)
	}
	return data, nil
}

func renderToolGuidance(toolNames []string, isSubAgent bool) string {
	if len(toolNames) == 0 {
		return ""
	}
	toolset := make(map[string]struct{}, len(toolNames))
	for _, n := range toolNames {
		toolset[n] = struct{}{}
	}
	has := func(name string) bool {
		_, ok := toolset[name]
		return ok
	}

	lines := make([]string, 0, 12)
	lines = append(lines, "- Session tool guidance is derived from allowed tools for this agent.")
	if has("view") {
		lines = append(lines, "- Read file contents with `view` before proposing edits.")
	}
	if has("edit") || has("multiedit") || has("write") {
		lines = append(lines, "- Prefer `edit`/`multiedit`/`write` for code changes instead of shell text replacement.")
	}
	if has("glob") || has("grep") || has("ls") {
		lines = append(lines, "- Use `glob`/`grep`/`ls` for discovery before changing files.")
	}
	if has("bash") {
		lines = append(lines, "- Use `bash` only for shell/system operations not covered by dedicated tools.")
	}
	if has("agent") {
		lines = append(lines, "- Delegate independent side tasks with `agent`; avoid duplicating delegated work.")
	}
	if has("todos") {
		lines = append(lines, "- Keep todos current while executing multi-step tasks.")
	}
	if has("web_search") || has("web_fetch") || has("fetch") {
		lines = append(lines, "- For web research, iterate focused queries and fetch primary sources before concluding.")
	}
	if isSubAgent {
		lines = append(lines, "- You are a subagent: execute directly; do not recursively delegate unless explicitly needed.")
	}
	return strings.Join(lines, "\n")
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func getGitStatus(ctx context.Context, dir string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: dir,
	})
	branch, err := getGitBranch(ctx, sh)
	if err != nil {
		return "", err
	}
	status, err := getGitStatusSummary(ctx, sh)
	if err != nil {
		return "", err
	}
	commits, err := getGitRecentCommits(ctx, sh)
	if err != nil {
		return "", err
	}
	return branch + status + commits, nil
}

func getGitBranch(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git branch --show-current 2>/dev/null")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", nil
	}
	return fmt.Sprintf("Current branch: %s\n", out), nil
}

func getGitStatusSummary(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git status --short 2>/dev/null | head -20")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "Status: clean\n", nil
	}
	return fmt.Sprintf("Status:\n%s\n", out), nil
}

func getGitRecentCommits(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git log --oneline -n 3 2>/dev/null")
	if err != nil || out == "" {
		return "", nil
	}
	out = strings.TrimSpace(out)
	return fmt.Sprintf("Recent commits:\n%s\n", out), nil
}

func (p *Prompt) Name() string {
	return p.name
}
