package main

import "fmt"

type DB interface {
	Insert(metrics) error
}

type Mysql struct {
}

func (m *Mysql) Insert(met metrics) error {
	fmt.Printf("Insert %+v", met)
	return nil
}
