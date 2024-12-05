package main

import (
	"fmt"
	"go-mail_indexer/backend/pkg/indexer"
	"log"
	"net/http"
	_ "net/http/pprof" // Importar pprof para habilitar el profiling

	"github.com/spf13/viper"
)

// func main() {
// 	// Ruta del archivo descargado y directorio de salida
// 	enronURL := "http://www.cs.cmu.edu/~enron/enron_mail_20110402.tgz"
// 	tgzPath := "enron_mail.tgz"
// 	outputDir := "enron_data"

// 	// 1. Descargar la base de datos
// 	fmt.Println("Descargando la base de datos de Enron...")
// 	if err := indexer.DownloadData(enronURL, tgzPath); err != nil {
// 		fmt.Printf("Error descargando dataset: %v\n", err)
// 		return
// 	}

// 	// 2. Extraer la base de datos
// 	fmt.Println("Extrayendo la base de datos de Enron...")
// 	if err := indexer.ExtractData(tgzPath, outputDir); err != nil {
// 		fmt.Printf("Error extrayendo dataset: %v\n", err)
// 		return
// 	}

// 	// 3. Indexar correos en ZincSearch
// 	fmt.Println("Indexando correos en ZincSearch...")
// 	if err := indexer.IndexData(outputDir); err != nil {
// 		fmt.Printf("Error indexando correos: %v\n", err)
// 		return
// 	}

// 	fmt.Println("¡Indexación completada!")
// }

func main() {
	// Iniciar servidor de pprof en segundo plano
	go func() {
		fmt.Println("Profiling habilitado en: http://localhost:6060/debug/pprof/")
		http.ListenAndServe("localhost:6060", nil)
	}()

	// Ruta del archivo descargado y directorio de salida
	tgzPath := "enron_mail.tgz"
	outputDir := "backend/data/enron_data"

	// Configurar Viper
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No se pudo leer el archivo .env: %v (se usarán variables del entorno si están configuradas)", err)
	}

	// Inicializar las configuraciones en indexer
	indexer.InitialConfig()

	// 1. Descargar la base de datos
	fmt.Println("Descargando la base de datos de Enron...")
	if err := indexer.DownloadData(indexer.EnronDataUrl, tgzPath); err != nil {
		fmt.Printf("Error descargando dataset: %v\n", err)
		return
	}

	// 2. Extraer la base de datos
	fmt.Println("Extrayendo la base de datos de Enron...")
	if err := indexer.ExtractData(tgzPath, outputDir); err != nil {
		fmt.Printf("Error extrayendo dataset: %v\n", err)
		return
	}

	// 3. Indexar correos en ZincSearch
	fmt.Println("Indexando correos en ZincSearch...")
	if err := indexer.IndexData(outputDir); err != nil {
		fmt.Printf("Error indexando correos: %v\n", err)
		return
	}

	fmt.Println("¡Indexación completada!")
}