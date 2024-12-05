package indexer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

const (
	outDataputdir = "enron_mail_data"
	zincIndex = "enron_mails"
)

// Variables configuradas dinámicamente
var (
	EnronDataUrl string
	ZincBaseUrl  string
	ZincUser     string
	ZincPassword string
)

type Email struct {
	Subject string `json:"subject"`
	Body string `json:"body"`
	From string `json:"from"`
	To string `json:"to"`
	Date string `json:"date"`
}

func InitialConfig() {
	EnronDataUrl = viper.GetString("ENRON_DATA_PATH")
	ZincBaseUrl = viper.GetString("ZINC_BASE_URL")
	ZincUser = viper.GetString("ZINC_FIRST_ADMIN_USER")
	ZincPassword = viper.GetString("ZINC_FIRST_ADMIN_PASSWORD")

	// Valores por defecto si las variables no están configuradas
	if EnronDataUrl == "" {
		EnronDataUrl = "https://www.cs.cmu.edu/~./enron/enron_mail_20150507.tar.gz"
	}
	if ZincBaseUrl == "" {
		ZincBaseUrl = "http://localhost:4080/api"
	}
	if ZincUser == "" || ZincPassword == "" {
		fmt.Println("Advertencia: Usuario o contraseña de ZincSearch no configurados. Usando valores por defecto.")
		ZincUser = "admin"
		ZincPassword = "admin123"
	}
}


func DownloadData(url, dest string) error {
	// Verificar si el archivo ya existe
	if fileInfo, err := os.Stat(dest); err == nil {
		// Consultar el tamaño del archivo en el servidor
		resp, err := http.Head(url)
		if err == nil {
			serverFileSize := resp.ContentLength
			if serverFileSize == fileInfo.Size() {
				fmt.Printf("El archivo %s ya está actualizado. No se descargará de nuevo.\n", dest)
				return nil
			}
		}
	}

	fmt.Println("Downloading data from: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractData extraer los datos del archivo tar.gz
func ExtractData(data, dataDir string) error {
	fmt.Println("Extracting data from: ", dataDir)
	file, err := os.Open(data)
	if err != nil {
		return err
	}
	defer file.Close() // cerrar el archivo al finalizar la función

	// crear un nuevo lector gzip para el archivo
	grz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer grz.Close()

	// crear un nuevo lector tar para el archivo gzip
	tarReader := tar.NewReader(grz)
	for {
		header, err := tarReader.Next() // obtener el siguiente encabezado del archivo
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		target := filepath.Join(dataDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
			file.Close()
		default:
			return fmt.Errorf("unknown type: %v in %s", header.Typeflag, header.Name)		
		}
	}

	return nil
}

// IndexData indexa los correos en ZincSearch
func IndexData(root string) error {
	// Verificar si el índice ya existe
	indexExists, err := checkZincIndex()
	if err != nil {
		return fmt.Errorf("error verificando el índice en ZincSearch: %w", err)
	}

	// Si el índice ya existe, evitar reindexar
	if indexExists {
		fmt.Println("El índice ya existe. No se realizará una nueva indexación.")
		return nil
	}

	// Recorrer los correos y enviarlos a ZincSearch
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".txt" {
			return nil
		}

		// Parsear el correo
		email, err := parseEmail(path)
		if err != nil {
			fmt.Printf("Error leyendo correo (%s): %v\n", path, err)
			return nil
		}

		// Indexar el correo
		if err := sendToZinc(email); err != nil {
			fmt.Printf("Error indexando correo (%s): %v\n", path, err)
			return nil
		}

		return nil
	})
}
// parseEmail lee un archivo de correo y devuelve un objeto Email
func parseEmail(path string) (*Email, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := string(content)
	subject := ""
	return &Email{
		Subject: subject,
		Body: lines,
		From: "",
		To: "",
		Date: "",
	}, nil
}
// checkZincIndex verifica si el índice ya existe en ZincSearch
func checkZincIndex() (bool, error) {
	url := fmt.Sprintf("%s/_index/%s", ZincBaseUrl, zincIndex)

	// Crear solicitud GET para verificar el índice
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error creando solicitud para ZincSearch: %w", err)
	}
	req.SetBasicAuth(ZincUser, ZincPassword)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error consultando ZincSearch: %w", err)
	}
	defer resp.Body.Close()

	// Verificar el estado de la respuesta
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("El índice %s ya existe en ZincSearch.\n", zincIndex)
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("El índice %s no existe en ZincSearch. Se creará al indexar.\n", zincIndex)
		return false, nil
	}

	// Otro error inesperado
	return false, fmt.Errorf("error inesperado de ZincSearch: %d", resp.StatusCode)
}
// sendToZinc envía un correo a ZincSearch
func sendToZinc(email *Email) error {
	// Crear cliente REST
	client := resty.New()
	client.SetBasicAuth(ZincUser, ZincPassword)// Autenticación básica

	// Indexar el correo en ZincSearch
	resp, err := client.R().
	SetHeader("Content-Type", "application/json").
	SetBody(email).
	Post(fmt.Sprintf("%s/%s/_doc", ZincBaseUrl, zincIndex))

	if err != nil {
		return fmt.Errorf("error indexando correo: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error indexando correo: %s", resp.Status())
	}

	return nil
}
