package plugin

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/devctl/pkg/runtime"
)

// Spec declares one plugin binding: an executable plus the seams it claims.
// Specs come from the book profile's plugins: section or from repeated
// --plugin seam=path CLI flags.
type Spec struct {
	ID    string            `yaml:"id" json:"id"`
	Path  string            `yaml:"path" json:"path"`
	Args  []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Env   map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Seams []string          `yaml:"seams" json:"seams"`
}

// Provenance identifies a running plugin for run metadata and page artifacts.
type Provenance struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Ops      []string       `json:"ops"`
	Declares map[string]any `json:"declares,omitempty"`
}

type managedPlugin struct {
	spec   Spec
	client runtime.Client
}

// Manager spawns the declared plugins, validates their handshakes against the
// seams they claim, and routes op calls. One plugin process per spec lives
// for the whole run; devctl's runtime client is safe for concurrent calls, so
// page workers share it.
type Manager struct {
	plugins []managedPlugin
	bySeam  map[string]*managedPlugin
}

// NewManager starts every spec and fail-fast validates that each claimed seam
// is (a) a seam the host knows and (b) advertised in the plugin's handshake
// capabilities. Binding resolution is first-wins per seam. On any error all
// started plugins are shut down.
func NewManager(ctx context.Context, specs []Spec, meta runtime.RequestMeta) (*Manager, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	factory := runtime.NewFactory(runtime.FactoryOptions{HandshakeTimeout: 5 * time.Second})
	m := &Manager{bySeam: map[string]*managedPlugin{}}
	for _, spec := range specs {
		if strings.TrimSpace(spec.Path) == "" {
			m.closeAll(ctx)
			return nil, fmt.Errorf("plugin %q: path is required", spec.ID)
		}
		if len(spec.Seams) == 0 {
			m.closeAll(ctx)
			return nil, fmt.Errorf("plugin %q: at least one seam is required", spec.ID)
		}
		client, err := factory.Start(ctx, runtime.PluginSpec{ID: spec.ID, Path: spec.Path, Args: spec.Args, Env: spec.Env}, runtime.StartOptions{Meta: meta})
		if err != nil {
			m.closeAll(ctx)
			return nil, fmt.Errorf("start plugin %q (%s): %w", spec.ID, spec.Path, err)
		}
		mp := &managedPlugin{spec: spec, client: client}
		m.plugins = append(m.plugins, *mp)
		for _, seam := range spec.Seams {
			if !knownSeam(seam) {
				m.closeAll(ctx)
				return nil, fmt.Errorf("plugin %q claims unknown seam %q (known: %s)", spec.ID, seam, strings.Join(KnownSeams, ", "))
			}
			if !client.SupportsOp(seam) {
				m.closeAll(ctx)
				return nil, fmt.Errorf("plugin %q claims seam %q but its handshake only advertises ops %v", spec.ID, seam, client.Handshake().Capabilities.Ops)
			}
			if _, taken := m.bySeam[seam]; !taken {
				m.bySeam[seam] = mp
			}
		}
	}
	return m, nil
}

func knownSeam(seam string) bool {
	return slices.Contains(KnownSeams, seam)
}

// Has reports whether some plugin is bound to the given seam. A nil Manager
// has no seams, so callers can hold a nil *Manager when no plugins are
// configured.
func (m *Manager) Has(seam string) bool {
	if m == nil {
		return false
	}
	_, ok := m.bySeam[seam]
	return ok
}

// Call dispatches an op to the plugin bound to that seam.
func (m *Manager) Call(ctx context.Context, seam string, input any, output any) error {
	if m == nil {
		return fmt.Errorf("no plugin manager configured")
	}
	mp, ok := m.bySeam[seam]
	if !ok {
		return fmt.Errorf("no plugin bound to seam %q", seam)
	}
	return mp.client.Call(ctx, seam, input, output)
}

// PluginIDFor names the plugin bound to a seam (for provenance strings).
func (m *Manager) PluginIDFor(seam string) string {
	if m == nil {
		return ""
	}
	if mp, ok := m.bySeam[seam]; ok {
		return mp.spec.ID
	}
	return ""
}

// Provenance describes every running plugin, sorted by ID.
func (m *Manager) Provenance() []Provenance {
	if m == nil {
		return nil
	}
	out := make([]Provenance, 0, len(m.plugins))
	for _, mp := range m.plugins {
		hs := mp.client.Handshake()
		out = append(out, Provenance{ID: mp.spec.ID, Name: hs.PluginName, Ops: append([]string(nil), hs.Capabilities.Ops...), Declares: hs.Declares})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Close shuts down every plugin process (SIGTERM, then SIGKILL of the process
// group after the runtime's shutdown timeout).
func (m *Manager) Close(ctx context.Context) {
	if m == nil {
		return
	}
	m.closeAll(ctx)
}

func (m *Manager) closeAll(ctx context.Context) {
	for _, mp := range m.plugins {
		_ = mp.client.Close(ctx)
	}
	m.plugins = nil
	m.bySeam = map[string]*managedPlugin{}
}

// ParseSeamBinding parses one --plugin flag value of the form seam=path.
// It returns a single-seam Spec whose ID derives from the executable name.
func ParseSeamBinding(value string) (Spec, error) {
	seam, path, ok := strings.Cut(value, "=")
	seam = strings.TrimSpace(seam)
	path = strings.TrimSpace(path)
	if !ok || seam == "" || path == "" {
		return Spec{}, fmt.Errorf("invalid --plugin value %q: expected seam=path (e.g. ocr.page=./my_plugin.py)", value)
	}
	if !knownSeam(seam) {
		return Spec{}, fmt.Errorf("unknown plugin seam %q (known: %s)", seam, strings.Join(KnownSeams, ", "))
	}
	id := seam + ":" + path
	return Spec{ID: id, Path: path, Seams: []string{seam}}, nil
}
