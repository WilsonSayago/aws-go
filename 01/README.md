# AWS Bedrock Go Example

Un ejemplo práctico de cómo usar Amazon Bedrock con Go para interactuar con modelos de IA generativa, específicamente Anthropic Claude.

## 📋 Descripción

Este proyecto demuestra cómo:
- Configurar el SDK de AWS para Go v2
- Autenticarse con AWS
- Invocar modelos de IA en Amazon Bedrock
- Manejar respuestas de modelos de lenguaje
- Implementar manejo de errores robusto

## 🚀 Requisitos Previos

- **Go 1.19+** instalado
- **Cuenta de AWS** activa
- **Credenciales de AWS** configuradas
- **Acceso a Amazon Bedrock** habilitado en tu cuenta

## 📦 Instalación

1. **Clonar el repositorio:**
   ```bash
   git clone <tu-repositorio>
   cd aws-go-01
   ```

2. **Instalar dependencias:**
   ```bash
   go mod tidy
   ```

## ⚙️ Configuración de AWS

### AWS CLI (Recomendado)
```bash
aws configure
```

### AWS SDK for Go v2

```bash
go get github.com/aws/aws-sdk-go-v2
```



## 🔑 Configuración de Permisos

Tu usuario de IAM necesita los siguientes permisos:

```json
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowListModels",
			"Effect": "Allow",
			"Action": [
				"bedrock:ListFoundationModels",
				"bedrock:GetFoundationModelAvailability"
			],
			"Resource": "*"
		},
		{
			"Sid": "AllowInvokeClaudeAndOthers",
			"Effect": "Allow",
			"Action": [
				"bedrock:InvokeModel"
			],
			"Resource": [
				"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-v2",
				"arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-v2:1",
				"arn:aws:bedrock:us-east-1::foundation-model/ai21.j2-mid-v1",
				"arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-text-lite-v1",
				"arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-text-express-v1"
			]
		}
	]
}
```

O usar la política gestionada: `AmazonBedrockFullAccess`

## 🎯 Habilitar Modelos en Bedrock

1. Ve a la **consola de Amazon Bedrock**
2. Selecciona **"Model access"** en la barra lateral
3. Habilita el modelo **`anthropic.claude-v2`**
4. Espera a que el estado cambie a **"Access granted"**

## 🏃‍♂️ Uso

### Ejecutar con región por defecto (us-east-1):
```bash
go run main.go
```

### Ejecutar con región específica:
```bash
go run main.go -region us-west-2
```

### Ejemplo de salida:
```
Using AWS region: us-east-1
Prompt:
 ¿Cuál es la capital de Francia?
Response from Anthropic Claude:
 La capital de Francia es París. París es la ciudad más grande de Francia y sirve como su centro político, económico y cultural.
```

## 🔧 Comandos Útiles de AWS CLI

### Listar modelos disponibles:
```bash
aws bedrock list-foundation-models --region us-east-1
```

### Verificar credenciales:
```bash
aws sts get-caller-identity
```

### Verificar acceso a Bedrock:
```bash
aws bedrock list-foundation-models --region us-east-1 --by-provider anthropic
```

## 📁 Estructura del Proyecto

```
aws-go-01/
├── main.go          # Aplicación principal
├── go.mod           # Dependencias de Go
├── go.sum           # Checksums de dependencias
└── README.md        # Este archivo
```

## 🛠️ Código Principal

El código incluye:

- **Estructuras tipadas** para requests y responses de Claude
- **Manejo de errores** específico para diferentes tipos de fallos
- **Configuración flexible** de región via flags
- **Formato correcto** de prompts para Anthropic Claude

## 🚨 Solución de Problemas

### Error: "no EC2 IMDS role found"
**Causa:** Credenciales de AWS no configuradas.
**Solución:** Configurar credenciales usando una de las opciones arriba.

### Error: "Could not resolve the foundation model"
**Causa:** Modelo no disponible en la región o no habilitado.
**Solución:** 
1. Verificar que el modelo esté disponible en tu región
2. Habilitar el modelo en la consola de Bedrock

### Error: "AccessDeniedException"
**Causa:** Permisos insuficientes.
**Solución:** Agregar permisos de Bedrock a tu usuario de IAM.

### Error: "no such host"
**Causa:** Bedrock no disponible en la región seleccionada.
**Solución:** Usar una región donde Bedrock esté disponible (us-east-1, us-west-2, etc.).

## 📚 Recursos Adicionales

- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Amazon Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [Anthropic Claude Model Parameters](https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html)
- [AWS Bedrock Pricing](https://aws.amazon.com/bedrock/pricing/)

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Por favor:

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -m 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🆘 Soporte

Si tienes problemas o preguntas:
1. Revisa la sección de **Solución de Problemas**
2. Consulta la documentación oficial de AWS
3. Abre un issue en este repositorio
