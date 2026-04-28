package syntax

import "testing"

func TestDefaultDialectIsEnabled(t *testing.T) {
	d := DefaultDialect()
	if !d.Enabled {
		t.Fatal("expected default dialect to be enabled")
	}
	if d.ID != DialectMSXBasicOfficial {
		t.Fatalf("expected default dialect %q, got %q", DialectMSXBasicOfficial, d.ID)
	}
}

func TestDialectOptionsContainsFutureDialectsDisabled(t *testing.T) {
	opts := DialectOptions()
	foundOfficial := false
	foundFuture := 0
	for _, opt := range opts {
		if opt.ID == DialectMSXBasicOfficial && opt.Enabled {
			foundOfficial = true
		}
		if (opt.ID == DialectMSXBas2ROM || opt.ID == DialectBasicDignified) && !opt.Enabled {
			foundFuture++
		}
	}
	if !foundOfficial {
		t.Fatal("expected enabled official MSX-BASIC option")
	}
	if foundFuture < 2 {
		t.Fatalf("expected disabled placeholders for future dialects, got %d", foundFuture)
	}
}
