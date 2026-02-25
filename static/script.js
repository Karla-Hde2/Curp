document.getElementById("curpForm").addEventListener("submit", async (e) => {
  e.preventDefault();

  // Tomar los datos del formulario
  const datos = Object.fromEntries(new FormData(e.target));

  try {
    // Enviar datos al backend
    const res = await fetch("/generar-curp", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(datos)
    });

    if (!res.ok) {
      const errorText = await res.text();
      document.getElementById("resultado").textContent = "Error: " + errorText;
      return;
    }

    const data = await res.json();

    // Mostrar CURP completa
    document.getElementById("resultado").textContent = "CURP: " + data.curp;

    // Mostrar desglose léxico
    const analisisDiv = document.getElementById("analisis");
    analisisDiv.innerHTML = "";
    for (const [pos, desc] of Object.entries(data.analisis)) {
      const p = document.createElement("p");
      p.textContent = desc;
      analisisDiv.appendChild(p);
    }
  } catch (err) {
    document.getElementById("resultado").textContent = "Error de conexión con el servidor";
  }
});