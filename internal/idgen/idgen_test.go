package idgen

import (
	"testing"
)

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateID()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateIDParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := GenerateID()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}