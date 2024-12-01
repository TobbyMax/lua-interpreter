package main

type Env struct {
	vars  map[string]float64
	stack []float64
}

func NewEnv() *Env {
	return &Env{
		vars:  make(map[string]float64),
		stack: []float64{},
	}
}

func (e *Env) Push(val float64) {
	e.stack = append(e.stack, val)
}

func (e *Env) Pop() float64 {
	if len(e.stack) == 0 {
		return 0
	}
	val := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return val
}
