package cipher

var _key = []byte("8pUsXuZw4z6B9EhGdKgNjQnjmVsYv2x5")

func GenerateKey(key string) {
	_key = []byte(key)
}

func XOR(src []byte) []byte {
	_klen := len(_key)
	for i := 0; i < len(src); i++ {
		src[i] ^= _key[i%_klen]
	}
	return src
}
