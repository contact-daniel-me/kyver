package navigation

func (e *Engine) Goto(symbol string, selectIndex int) (*GotoResult, error) {
	candidates, selected, err := e.resolveSymbol(symbol, selectIndex)
	if err != nil {
		return nil, err
	}

	res := &GotoResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
	} else {
		res.Symbol = selected
	}

	return res, nil
}
