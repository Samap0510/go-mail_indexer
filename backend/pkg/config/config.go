package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	EnronDataUrl string
	ZincBaseURL string
	ZincUser string
	ZincPassword string
}

// LoadConfig carga las configuraciones desde .env o variables de entorno
func LoadConfig() (*Config, error) {
	// Configurar Viper
	viper.SetConfigFile(".env") // Archivo .env
	viper.AutomaticEnv()        // Sobrescribir con variables del entorno del sistema

	// Leer el archivo .env
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No se pudo leer el archivo .env: %v (se usarán variables del entorno si están configuradas)", err)
	}

	// Asignar valores
	config := &Config{
		ZincBaseURL:  viper.GetString("ZINC_BASE_URL"),
		ZincUser:     viper.GetString("ZINC_FIRST_ADMIN_USER"),
		ZincPassword: viper.GetString("ZINC_FIRST_ADMIN_PASSWORD"),
		EnronDataUrl: viper.GetString("ENRON_DATA_PATH"),
	}

	// Validar valores requeridos
	if config.ZincBaseURL == "" || config.ZincUser == "" || config.ZincPassword == "" {
		log.Fatal("Faltan configuraciones obligatorias: ZINC_BASE_URL, ZINC_FIRST_ADMIN_USER o ZINC_FIRST_ADMIN_PASSWORD")
	}
	return config, nil
}