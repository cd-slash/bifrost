package handlers

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
	configstoreTables "github.com/maximhq/bifrost/framework/configstore/tables"
	"github.com/maximhq/bifrost/plugins/governance"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

type handlerTestLogger struct{}

func (l *handlerTestLogger) Debug(msg string, args ...any)                     {}
func (l *handlerTestLogger) Info(msg string, args ...any)                      {}
func (l *handlerTestLogger) Warn(msg string, args ...any)                      {}
func (l *handlerTestLogger) Error(msg string, args ...any)                     {}
func (l *handlerTestLogger) Fatal(msg string, args ...any)                     {}
func (l *handlerTestLogger) SetLevel(level schemas.LogLevel)                   {}
func (l *handlerTestLogger) SetOutputType(outputType schemas.LoggerOutputType) {}

type mockGovernanceManager struct {
	reloadProfilesCalls int
}

type contextSafeStore struct {
	configstore.ConfigStore
}

func (s *contextSafeStore) GetRoutingProfiles(ctx context.Context) ([]configstoreTables.TableRoutingProfile, error) {
	return s.ConfigStore.GetRoutingProfiles(context.Background())
}

func (s *contextSafeStore) CreateRoutingProfile(ctx context.Context, profile *configstoreTables.TableRoutingProfile, tx ...*gorm.DB) error {
	return s.ConfigStore.CreateRoutingProfile(context.Background(), profile, tx...)
}

func (s *contextSafeStore) UpdateRoutingProfile(ctx context.Context, profile *configstoreTables.TableRoutingProfile, tx ...*gorm.DB) error {
	return s.ConfigStore.UpdateRoutingProfile(context.Background(), profile, tx...)
}

func (s *contextSafeStore) DeleteRoutingProfile(ctx context.Context, id string, tx ...*gorm.DB) error {
	return s.ConfigStore.DeleteRoutingProfile(context.Background(), id, tx...)
}

func (s *contextSafeStore) GetPlugin(ctx context.Context, name string) (*configstoreTables.TablePlugin, error) {
	return s.ConfigStore.GetPlugin(context.Background(), name)
}

func (s *contextSafeStore) CreatePlugin(ctx context.Context, plugin *configstoreTables.TablePlugin, tx ...*gorm.DB) error {
	return s.ConfigStore.CreatePlugin(context.Background(), plugin, tx...)
}

func (s *contextSafeStore) UpdatePlugin(ctx context.Context, plugin *configstoreTables.TablePlugin, tx ...*gorm.DB) error {
	return s.ConfigStore.UpdatePlugin(context.Background(), plugin, tx...)
}

func (s *contextSafeStore) GetProviders(ctx context.Context) ([]configstoreTables.TableProvider, error) {
	return s.ConfigStore.GetProviders(context.Background())
}

func (m *mockGovernanceManager) GetGovernanceData() *governance.GovernanceData { return nil }
func (m *mockGovernanceManager) ReloadVirtualKey(ctx context.Context, id string) (*configstoreTables.TableVirtualKey, error) {
	return nil, nil
}
func (m *mockGovernanceManager) RemoveVirtualKey(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) ReloadTeam(ctx context.Context, id string) (*configstoreTables.TableTeam, error) {
	return nil, nil
}
func (m *mockGovernanceManager) RemoveTeam(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) ReloadCustomer(ctx context.Context, id string) (*configstoreTables.TableCustomer, error) {
	return nil, nil
}
func (m *mockGovernanceManager) RemoveCustomer(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) ReloadModelConfig(ctx context.Context, id string) (*configstoreTables.TableModelConfig, error) {
	return nil, nil
}
func (m *mockGovernanceManager) RemoveModelConfig(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) ReloadProvider(ctx context.Context, provider schemas.ModelProvider) (*configstoreTables.TableProvider, error) {
	return nil, nil
}
func (m *mockGovernanceManager) RemoveProvider(ctx context.Context, provider schemas.ModelProvider) error {
	return nil
}
func (m *mockGovernanceManager) ReloadRoutingRule(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) RemoveRoutingRule(ctx context.Context, id string) error { return nil }
func (m *mockGovernanceManager) ReloadRoutingProfiles(ctx context.Context) error {
	m.reloadProfilesCalls++
	return nil
}

func createHandlerTestStore(t *testing.T) configstore.ConfigStore {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "handlers-test-config.db")
	store, err := configstore.NewConfigStore(context.Background(), &configstore.Config{
		Enabled: true,
		Type:    configstore.ConfigStoreTypeSQLite,
		Config:  &configstore.SQLiteConfig{Path: dbPath},
	}, &handlerTestLogger{})
	if err != nil {
		t.Fatalf("failed to create config store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close(context.Background())
	})
	return &contextSafeStore{ConfigStore: store}
}

func newTestRequestCtx(body string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fasthttp.MethodPost)
	ctx.Request.SetBodyString(body)
	ctx.Request.Header.SetContentType("application/json")
	return ctx
}

func parseJSONBody(t *testing.T, ctx *fasthttp.RequestCtx) map[string]any {
	t.Helper()
	out := map[string]any{}
	if err := json.Unmarshal(ctx.Response.Body(), &out); err != nil {
		t.Fatalf("failed to decode JSON response: %v, body=%s", err, string(ctx.Response.Body()))
	}
	return out
}

func TestRoutingProfilesCRUDHandlers(t *testing.T) {
	SetLogger(&handlerTestLogger{})
	store := createHandlerTestStore(t)
	mgr := &mockGovernanceManager{}
	h, err := NewGovernanceHandler(mgr, store)
	if err != nil {
		t.Fatalf("failed to create governance handler: %v", err)
	}

	createCtx := newTestRequestCtx(`{"id":"rp1","name":"Light","virtual_provider":"light","enabled":true,"strategy":"ordered_failover","targets":[{"provider":"cerebras","virtual_model":"light","model":"glm-4.7-flash","priority":1,"enabled":true}]}`)
	h.createRoutingProfile(createCtx)
	if createCtx.Response.StatusCode() != fasthttp.StatusCreated {
		t.Fatalf("expected create 201, got %d body=%s", createCtx.Response.StatusCode(), string(createCtx.Response.Body()))
	}

	listCtx := &fasthttp.RequestCtx{}
	h.getRoutingProfiles(listCtx)
	if listCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected list 200, got %d", listCtx.Response.StatusCode())
	}
	listBody := parseJSONBody(t, listCtx)
	if int(listBody["count"].(float64)) != 1 {
		t.Fatalf("expected 1 profile after create, got %+v", listBody)
	}

	getCtx := &fasthttp.RequestCtx{}
	getCtx.SetUserValue("profile_id", "rp1")
	h.getRoutingProfile(getCtx)
	if getCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected detail 200, got %d", getCtx.Response.StatusCode())
	}

	updateCtx := newTestRequestCtx(`{"name":"Light Updated"}`)
	updateCtx.SetUserValue("profile_id", "rp1")
	h.updateRoutingProfile(updateCtx)
	if updateCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected update 200, got %d body=%s", updateCtx.Response.StatusCode(), string(updateCtx.Response.Body()))
	}

	deleteCtx := &fasthttp.RequestCtx{}
	deleteCtx.SetUserValue("profile_id", "rp1")
	h.deleteRoutingProfile(deleteCtx)
	if deleteCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected delete 200, got %d body=%s", deleteCtx.Response.StatusCode(), string(deleteCtx.Response.Body()))
	}

	listAfterCtx := &fasthttp.RequestCtx{}
	h.getRoutingProfiles(listAfterCtx)
	listAfterBody := parseJSONBody(t, listAfterCtx)
	if int(listAfterBody["count"].(float64)) != 0 {
		t.Fatalf("expected 0 profiles after delete, got %+v", listAfterBody)
	}

	if mgr.reloadProfilesCalls < 3 {
		t.Fatalf("expected reload routing profiles to be called on create/update/delete, got %d", mgr.reloadProfilesCalls)
	}
}

func TestRoutingProfilesImportExportAndSimulateHandlers(t *testing.T) {
	SetLogger(&handlerTestLogger{})
	store := createHandlerTestStore(t)
	mgr := &mockGovernanceManager{}
	h, err := NewGovernanceHandler(mgr, store)
	if err != nil {
		t.Fatalf("failed to create governance handler: %v", err)
	}

	importCtx := newTestRequestCtx(`{"routing_profiles":[{"id":"rp-light","name":"Light","virtual_provider":"light","enabled":true,"strategy":"ordered_failover","targets":[{"provider":"anthropic","virtual_model":"light","model":"claude-3-5-haiku-latest","priority":1,"enabled":true},{"provider":"cerebras","model":"glm-4.7-flash","priority":2,"enabled":true}]}]}`)
	h.importRoutingProfiles(importCtx)
	if importCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected import 200, got %d body=%s", importCtx.Response.StatusCode(), string(importCtx.Response.Body()))
	}

	exportCtx := &fasthttp.RequestCtx{}
	h.exportRoutingProfiles(exportCtx)
	if exportCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected export 200, got %d", exportCtx.Response.StatusCode())
	}
	exportBody := parseJSONBody(t, exportCtx)
	plugin, ok := exportBody["plugin"].(map[string]any)
	if !ok {
		t.Fatalf("expected plugin object in export response, got %+v", exportBody)
	}
	config, ok := plugin["config"].(map[string]any)
	if !ok {
		t.Fatalf("expected plugin config in export response")
	}
	if len(config["routing_profiles"].([]any)) != 1 {
		t.Fatalf("expected 1 exported routing profile")
	}

	simulateCtx := newTestRequestCtx(`{"model":"light/light","request_type":"chat","capabilities":["text"]}`)
	h.simulateRoutingProfile(simulateCtx)
	if simulateCtx.Response.StatusCode() != fasthttp.StatusOK {
		t.Fatalf("expected simulate 200, got %d body=%s", simulateCtx.Response.StatusCode(), string(simulateCtx.Response.Body()))
	}
	simBody := parseJSONBody(t, simulateCtx)
	if simBody["primary"].(string) != "anthropic/claude-3-5-haiku-latest" {
		t.Fatalf("unexpected simulation primary: %+v", simBody)
	}
}
