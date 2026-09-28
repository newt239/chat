//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	err := entc.Generate("./schema", &gen.Config{
		// 集計クエリを生 SQL で書くため ExecQuery を有効にする
		Features: []gen.Feature{gen.FeatureExecQuery, gen.FeatureUpsert},
	})
	if err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
