package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Cliente struct {
	IDCliente        int
	Edad             int
	IngresosAnuales  int
	PuntuacionCompra int
}

func cargarDatos(url string) ([]Cliente, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = 4

	var clientes []Cliente
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		id, _ := strconv.Atoi(record[0])
		edad, _ := strconv.Atoi(record[1])
		ingresos, _ := strconv.Atoi(record[2])
		puntuacion, _ := strconv.Atoi(record[3])

		clientes = append(clientes, Cliente{id, edad, ingresos, puntuacion})
	}

	return clientes, nil
}

func distancia(a, b Cliente) float64 {
	return math.Sqrt(float64((a.Edad-b.Edad)*(a.Edad-b.Edad) + (a.IngresosAnuales-b.IngresosAnuales)*(a.IngresosAnuales-b.IngresosAnuales) + (a.PuntuacionCompra-b.PuntuacionCompra)*(a.PuntuacionCompra-b.PuntuacionCompra)))
}

func asignarACentroides(clientes []Cliente, centroides []Cliente) []int {
	asignaciones := make([]int, len(clientes))
	for i, c := range clientes {
		menorDistancia := math.MaxFloat64
		for j, centroide := range centroides {
			dist := distancia(c, centroide)
			if dist < menorDistancia {
				menorDistancia = dist
				asignaciones[i] = j
			}
		}
	}
	return asignaciones
}

func recalcularCentroides(clientes []Cliente, asignaciones []int, k int) []Cliente {
	centroides := make([]Cliente, k)
	contadores := make([]int, k)

	for i, c := range clientes {
		indice := asignaciones[i]
		centroides[indice].Edad += c.Edad
		centroides[indice].IngresosAnuales += c.IngresosAnuales
		centroides[indice].PuntuacionCompra += c.PuntuacionCompra
		contadores[indice]++
	}

	for i := range centroides {
		if contadores[i] != 0 {
			centroides[i].Edad /= contadores[i]
			centroides[i].IngresosAnuales /= contadores[i]
			centroides[i].PuntuacionCompra /= contadores[i]
		}
	}

	return centroides
}

func KMeansConcurrente(clientes []Cliente, k, numGoroutines int) []Cliente {
	partSize := len(clientes) / numGoroutines
	centroides := make([]Cliente, k)

	canal := make(chan []Cliente)

	for i := 0; i < numGoroutines; i++ {
		inicio := i * partSize
		fin := inicio + partSize
		if i == numGoroutines-1 {
			fin = len(clientes)
		}

		go func(inicio, fin int) {
			subClientes := clientes[inicio:fin]
			subAsignaciones := asignarACentroides(subClientes, centroides)
			subCentroides := recalcularCentroides(subClientes, subAsignaciones, k)
			canal <- subCentroides
		}(inicio, fin)
	}

	for i := 0; i < numGoroutines; i++ {
		subCentroides := <-canal
		for j := range centroides {
			centroides[j].Edad += subCentroides[j].Edad
			centroides[j].IngresosAnuales += subCentroides[j].IngresosAnuales
			centroides[j].PuntuacionCompra += subCentroides[j].PuntuacionCompra
		}
	}

	for i := range centroides {
		centroides[i].Edad /= numGoroutines
		centroides[i].IngresosAnuales /= numGoroutines
		centroides[i].PuntuacionCompra /= numGoroutines
	}

	return centroides
}

func main() {
	clientes, err := cargarDatos("https://raw.githubusercontent.com/JBenites10/TF_Programacion_Concurrente/main/Dataset.csv")
	if err != nil {
		log.Fatal(err)
	}

	k := 3
	numGoroutines := 6
	numEjecuciones := 1000

	var mejorCentroides []Cliente
	menorDistancia := math.MaxFloat64

	for i := 0; i < numEjecuciones; i++ {
		centroides := KMeansConcurrente(clientes, k, numGoroutines)
		asignaciones := asignarACentroides(clientes, centroides)
		distanciaTotal := 0.0
		for j, c := range clientes {
			distanciaTotal += distancia(c, centroides[asignaciones[j]])
		}
		if distanciaTotal < menorDistancia {
			menorDistancia = distanciaTotal
			mejorCentroides = centroides
		}
	}

	fmt.Println("Mejores centroides:", mejorCentroides)
}
