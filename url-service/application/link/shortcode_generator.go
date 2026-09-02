package link

type ShortCodeGenerator interface {
	GenerateRandom(length int) (string, error)
	GenerateFromHash(url string) (string, error)
	GenerateFromID(id int64) (string, error)
	GenerateCustom(slug string) (string, error)
}