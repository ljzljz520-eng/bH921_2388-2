package store_test

import (
	"path/filepath"
	"testing"

	"weddinglive/internal/domain"
	"weddinglive/internal/store"
)

func TestJSONStorePersistsState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "wedding-live.json")
	repository, err := store.Open(path, domain.State{})
	if err != nil {
		t.Fatal(err)
	}
	err = repository.Update(func(state *domain.State) error {
		state.Accounts = append(state.Accounts, domain.Account{ID: "acct-001", Name: "固定摄影师", Token: "fixed-token"})
		state.NextAccountID = 2
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	reloaded, err := store.Open(path, domain.State{})
	if err != nil {
		t.Fatal(err)
	}
	state := reloaded.Snapshot()
	if state.NextAccountID != 2 || len(state.Accounts) != 1 || state.Accounts[0].Token != "fixed-token" {
		t.Fatalf("unexpected reloaded state: %+v", state)
	}
}
