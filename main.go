package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"unicode"
)

type Resultado struct {
	Curp   string
	Lexico []string
	Error  string
}

var tpl = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Generador y Análisis de CURP</title>
    <style>
        body { font-family: Arial; display: flex; }
        .formulario { width: 40%; padding: 20px; }
        .resultado { width: 60%; padding: 20px; background: #f9f9f9; }
        h2 { color: #333; }
        input { margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="formulario">
        <h2>Generador de CURP</h2>
        <form method="POST" action="/generar">
            Nombre: <input type="text" name="nombre"><br>
            Apellido paterno: <input type="text" name="paterno"><br>
            Apellido materno: <input type="text" name="materno"><br>
            Año de nacimiento (AAAA): <input type="text" name="anio"><br>
            Mes de nacimiento (MM): <input type="text" name="mes"><br>
            Día de nacimiento (DD): <input type="text" name="dia"><br>
            Sexo (H/M): <input type="text" name="sexo"><br>
            Estado (OC=Oaxaca, TL=Tlaxcala): <input type="text" name="estado"><br>
            <input type="submit" value="Generar CURP">
        </form>
    </div>
    <div class="resultado">
        {{if .Error}}
            <h3 style="color:red;">Error: {{.Error}}</h3>
        {{else if .Curp}}
            <h2>CURP generada: {{.Curp}}</h2>
            <h3>Desglose Léxico de la CURP</h3>
            <ul>
                {{range .Lexico}}
                    <li>{{.}}</li>
                {{end}}
            </ul>
        {{end}}
    </div>
</body>
</html>
`))

func formHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func generarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	nombre := strings.ToUpper(r.FormValue("nombre"))
	paterno := strings.ToUpper(r.FormValue("paterno"))
	materno := strings.ToUpper(r.FormValue("materno"))
	anio := r.FormValue("anio")
	mes := r.FormValue("mes")
	dia := r.FormValue("dia")
	sexo := strings.ToUpper(r.FormValue("sexo"))
	estado := strings.ToUpper(r.FormValue("estado"))

	// Validaciones de texto
	if !validarTexto(nombre) || !validarTexto(paterno) || !validarTexto(materno) {
		tpl.Execute(w, Resultado{Error: "Nombres y apellidos deben contener solo letras y mínimo 3 caracteres"})
		return
	}
	if contieneObscenidad(nombre) || contieneObscenidad(paterno) || contieneObscenidad(materno) {
		tpl.Execute(w, Resultado{Error: "Los nombres o apellidos contienen palabras no permitidas"})
		return
	}

	// Validar estado
	if estado != "OC" && estado != "TL" {
		tpl.Execute(w, Resultado{Error: "Solo se permiten CURPs de Oaxaca (OC) y Tlaxcala (TL)"})
		return
	}

	// Construir fecha
	fechaStr := fmt.Sprintf("%s-%s-%s", anio, mes, dia)
	fecha, err := time.Parse("2006-01-02", fechaStr)
	if err != nil {
		tpl.Execute(w, Resultado{Error: "Fecha inválida"})
		return
	}

	// Validar edad máxima
	hoy := time.Now()
	edad := hoy.Year() - fecha.Year()
	if hoy.YearDay() < fecha.YearDay() {
		edad--
	}
	if edad > 120 {
		tpl.Execute(w, Resultado{Error: "Edad mayor a 120 años, CURP no válida"})
		return
	}

	// Validar 29 de febrero
	if fecha.Month() == time.February && fecha.Day() == 29 {
		if !esBisiesto(fecha.Year()) {
			tpl.Execute(w, Resultado{Error: fmt.Sprintf("El año %d no es bisiesto, fecha inválida", fecha.Year())})
			return
		}
	}

	// Construcción simplificada de CURP
	curp := fmt.Sprintf("%s%s%s%s%s%s%s%s",
		string(paterno[0]),
		buscarVocalInterna(paterno),
		string(materno[0]),
		string(nombre[0]),
		fecha.Format("060102"), // AAMMDD
		sexo,
		estado,
		"01", // homoclave simplificada
	)

	// Análisis léxico
	lexico := []string{
		fmt.Sprintf("%s → Inicial del apellido paterno", string(paterno[0])),
		fmt.Sprintf("%s → Primera vocal interna del apellido paterno", buscarVocalInterna(paterno)),
		fmt.Sprintf("%s → Inicial del apellido materno", string(materno[0])),
		fmt.Sprintf("%s → Inicial del nombre", string(nombre[0])),
		fmt.Sprintf("%s → Fecha de nacimiento (AAMMDD)", fecha.Format("060102")),
		fmt.Sprintf("%s → Sexo (H=Hombre, M=Mujer)", sexo),
		fmt.Sprintf("%s → Estado de nacimiento (OC=Oaxaca, TL=Tlaxcala)", estado),
		"01 → Homoclave simplificada",
	}

	tpl.Execute(w, Resultado{Curp: curp, Lexico: lexico})
}

func esBisiesto(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}

func buscarVocalInterna(paterno string) string {
	vocales := "AEIOU"
	for _, c := range paterno[1:] {
		if strings.ContainsRune(vocales, rune(c)) {
			return string(c)
		}
	}
	return "X"
}

func validarTexto(texto string) bool {
	if len(texto) < 3 {
		return false
	}
	for _, c := range texto {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

func contieneObscenidad(texto string) bool {
	obscenas := []string{"PUTA", "PENE", "MAMADA", "VERGA", "CULO"} // lista básica
	for _, palabra := range obscenas {
		if strings.Contains(texto, palabra) {
			return true
		}
	}
	return false
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/generar", generarHandler)
	fmt.Println("Servidor corriendo en http://localhost:3030/")
	http.ListenAndServe(":3030", nil)
}