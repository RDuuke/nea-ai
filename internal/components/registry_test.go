package components

import (
	"testing"

	"nea-ai/internal/model"
)

func TestDefaultRegistryIncludesCoreComponents(t *testing.T) {
	registry := DefaultRegistry()

	for _, id := range []model.ComponentID{model.ComponentBrain, model.ComponentFlow} {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("expected component %q to be registered", id)
		}
	}
}

func TestDefaultRegistryRejectsUnknownComponent(t *testing.T) {
	registry := DefaultRegistry()

	if _, ok := registry.Get(model.ComponentID("unknown")); ok {
		t.Fatal("expected unknown component to be absent")
	}
}
