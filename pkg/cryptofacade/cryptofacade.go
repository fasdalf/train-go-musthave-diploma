package cryptofacade

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(body, key string) string {
	// создаём новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	// передаём байты для хеширования
	h.Write([]byte(body))
	// получаем хеш в виде строки
	return base64.StdEncoding.EncodeToString(h.Sum([]byte(key)))
}
