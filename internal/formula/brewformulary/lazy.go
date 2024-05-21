package brewformulary

// // NewLazy creates a pre-loaded formulary
// func NewLazy(client *brewclient.Client, maxGoroutines int) (formula.ConcurrentFormulary, error) {
// 	return &Lazy{
// 		client:        client,
// 		loaded:        map[string]*v1.Info{},
// 		maxGoroutines: maxGoroutines,
// 	}, nil
// }

// // Lazy is a formulary that lazy-loads metadata from the Homebrew API
// type Lazy struct {
// 	client        *brewclient.Client
// 	loaded        map[string]*v1.Info // map of visited formulae
// 	maxGoroutines int
// }

// // Fetch implements formula.Formulary.
// func (f *Lazy) Fetch(ctx context.Context, name string) (formula.MultiPlatformFormula, error) {
// 	return f.fetch(ctx, name)
// }

// // FetchAll implements formula.ConcurrentFormulary.
// func (f *Lazy) FetchAll(ctx context.Context, names []string) ([]formula.MultiPlatformFormula, error) {
// 	fetchers := iter.Mapper[string, formula.MultiPlatformFormula]{MaxGoroutines: f.maxGoroutines}
// 	return fetchers.MapErr(names, func(namep *string) (formula.MultiPlatformFormula, error) {
// 		return f.fetch(ctx, *namep)
// 	})
// }

// // fetch fetches general metadata
// func (f *Lazy) fetch(_ context.Context, name string) (formula.MultiPlatformFormula, error) {
// 	panic("not implemented")
// 	// data := f.index.Find(name)
// 	// if data == nil {
// 	// 	return nil, brew.NewErrFormulaNotFound(name)
// 	// }
// 	// return formula.FromV1(data), nil
// }
