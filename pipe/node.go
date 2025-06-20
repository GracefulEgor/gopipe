package pipe

import "context"

type Node[I any, O any] interface {
	In() []<- chan I 
	Out() []chan<- O  
	Run(ctx context.Context) error
}