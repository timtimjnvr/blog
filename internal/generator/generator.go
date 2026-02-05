package generator

type GenerateAndValidate interface {
	Generate() error
	Validate() error
}
