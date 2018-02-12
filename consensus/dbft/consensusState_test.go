package dbft

import (
	"testing"
)

func TestConsensusState_HasFlag(t *testing.T) {
	type args struct {
		flag ConsensusState
	}
	tests := []struct {
		name  string
		state ConsensusState
		args  args
		want  bool
	}{
		{
			name:"test",
			state:Primary,
			args:args{flag:Primary},
			want:true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.HasFlag(tt.args.flag); got != tt.want {
				t.Errorf("ConsensusState.HasFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
