package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    var wg sync.WaitGroup
    clienteChan := make(chan Cliente, 10)
    clientes := make([]Cliente, 0, 1000000)

    go func() {
        for c := range clienteChan {
            clientes = append(clientes, c)
        }
    }()

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c := Cliente{
                IDCliente:        fmt.Sprintf("ID%v", rand.Int()),
                Edad:             rand.Intn(65) + 18,
                IngresosAnuales:  float64(rand.Intn(100000)),
                PuntuacionCompra: rand.Intn(100) + 1, 
            }
            clienteChan <- c
        }()
    }
    wg.Wait()
    close(clienteChan)
    fmt.Println("Total de clientes generados:", len(clientes))
}