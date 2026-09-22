package dataplane

import (
	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

type RuntimeState struct {
	Config *controlv1.ConfigSnapshot
}

func BuildRuntimeState(
	snapshot *controlv1.ConfigSnapshot,
) *RuntimeState {
	return &RuntimeState{
		Config: snapshot,
	}
}
