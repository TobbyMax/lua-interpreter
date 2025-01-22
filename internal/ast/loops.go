package ast

type (
	While struct {
		Exp   Expression
		Block Block
	}
	Repeat struct {
		Block Block
		Exp   Expression
	}
	For struct {
		Name  string
		Init  Expression
		Limit Expression
		Step  *Expression
		Block Block
	}
	ForIn struct {
		Names []string
		Exps  []Expression
		Block Block
	}
)
