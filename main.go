package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Datos struct {
	Nombre     string `json:"nombre"`
	ApellidoP  string `json:"apellido_p"`
	ApellidoM  string `json:"apellido_m"`
	FechaNac   string `json:"fecha_nac"`
	Sexo       string `json:"sexo"`
	Entidad    string `json:"entidad"`
}

// --- Validaciones léxicas ---
func validarTokens(d Datos) error {
	reNombre := regexp.MustCompile(`^[A-Za-zÁÉÍÓÚÑ]+$`)
	if !reNombre.MatchString(d.Nombre) || !reNombre.MatchString(d.ApellidoP) || !reNombre.MatchString(d.ApellidoM) {
		return fmt.Errorf("Nombre o apellidos inválidos")
	}
	reFecha := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !reFecha.MatchString(d.FechaNac) {
		return fmt.Errorf("Fecha inválida")
	}
	if d.Sexo != "H" && d.Sexo != "M" {
		return fmt.Errorf("Sexo inválido")
	}
	if d.Entidad != "OAX" && d.Entidad != "TLX" {
		return fmt.Errorf("Entidad inválida")
	}
	return nil
}

// --- Validaciones sintácticas ---
func validarSintaxis(d Datos) error {
	if d.Nombre == "" || d.ApellidoP == "" || d.ApellidoM == "" || d.FechaNac == "" || d.Sexo == "" || d.Entidad == "" {
		return fmt.Errorf("Faltan campos obligatorios")
	}
	return nil
}

// --- Generador de CURP ---
func generarCURP(d Datos) string {
	nombre := strings.ToUpper(d.Nombre)
	apP := strings.ToUpper(d.ApellidoP)
	apM := strings.ToUpper(d.ApellidoM)

	parte1 := apP[:1] + primeraVocalInterna(apP) + apM[:1] + nombre[:1]
	parte2 := d.FechaNac[2:4] + d.FechaNac[5:7] + d.FechaNac[8:10]
	parte3 := d.Sexo
	parte4 := d.Entidad
	parte5 := primeraConsonanteInterna(apP) + primeraConsonanteInterna(apM) + primeraConsonanteInterna(nombre)
	parte6 := "01"

	return parte1 + parte2 + parte3 + parte4 + parte5 + parte6
}

func primeraVocalInterna(s string) string {
	vocales := "AEIOU"
	for _, c := range s[1:] {
		if strings.ContainsRune(vocales, c) {
			return string(c)
		}
	}
	return "X"
}

func primeraConsonanteInterna(s string) string {
	consonantes := "BCDFGHJKLMNÑPQRSTVWXYZ"
	for _, c := range s[1:] {
		if strings.ContainsRune(consonantes, c) {
			return string(c)
		}
	}
	return "X"
}

// --- Analizador léxico de CURP ---
func analizarCURP(curp string) map[int]string {
	resultado := make(map[int]string)
	if len(curp) < 18 {
		resultado[0] = "CURP incompleta"
		return resultado
	}
	resultado[0] = fmt.Sprintf("%c: Inicial del primer apellido", curp[0])
	resultado[1] = fmt.Sprintf("%c: Primera vocal interna del primer apellido", curp[1])
	resultado[2] = fmt.Sprintf("%c: Inicial del segundo apellido", curp[2])
	resultado[3] = fmt.Sprintf("%c: Inicial del nombre", curp[3])
	resultado[4] = fmt.Sprintf("%c: Penúltimo dígito del año de nacimiento", curp[4])
	resultado[5] = fmt.Sprintf("%c: Último dígito del año de nacimiento", curp[5])
	resultado[6] = fmt.Sprintf("%c: Primer dígito del mes de nacimiento", curp[6])
	resultado[7] = fmt.Sprintf("%c: Segundo dígito del mes de nacimiento", curp[7])
	resultado[8] = fmt.Sprintf("%c: Primer dígito del día de nacimiento", curp[8])
	resultado[9] = fmt.Sprintf("%c: Segundo dígito del día de nacimiento", curp[9])
	resultado[10] = fmt.Sprintf("%c: Sexo (H/M)", curp[10])
	resultado[11] = fmt.Sprintf("%c: Primera letra de la entidad federativa", curp[11])
	resultado[12] = fmt.Sprintf("%c: Segunda letra de la entidad federativa", curp[12])
	resultado[13] = fmt.Sprintf("%c: Primera consonante interna del primer apellido", curp[13])
	resultado[14] = fmt.Sprintf("%c: Primera consonante interna del segundo apellido", curp[14])
	resultado[15] = fmt.Sprintf("%c: Primera consonante interna del nombre", curp[15])
	resultado[16] = fmt.Sprintf("%c: Dígito de homoclave", curp[16])
	resultado[17] = fmt.Sprintf("%c: Letra/número de homoclave", curp[17])
	return resultado
}

func handlerGenerarCURP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var datos Datos
		json.NewDecoder(r.Body).Decode(&datos)

		if err := validarTokens(datos); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validarSintaxis(datos); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		curp := generarCURP(datos)
		analisis := analizarCURP(curp)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"curp":     curp,
			"analisis": analisis,
		})
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/generar-curp", handlerGenerarCURP)
	fmt.Println("Servidor corriendo en http://localhost:3030")
	http.ListenAndServe(":3000", nil)
}