package link

type ShortCodeGenerator interface {
	GenerateUnique(sequence int64, randomLength int) (string, error)
}