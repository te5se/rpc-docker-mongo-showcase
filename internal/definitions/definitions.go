package definitions

type CatFactReader interface {
	GetCatFacts(page int, pageSize int) ([]string, error)
}
