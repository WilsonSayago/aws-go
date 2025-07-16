# AWS RAG System - Bedrock + S3

Sistema RAG (Retrieval Augmented Generation) que combina AWS Bedrock con S3 para procesamiento de documentos y generación de respuestas contextuales.

## 📋 Descripción

Este proyecto implementa un sistema RAG completo que:
- **Lee documentos** existentes desde Amazon S3
- **Procesa y divide** el contenido en chunks manejables
- **Busca información relevante** basándose en consultas del usuario
- **Genera respuestas** usando Amazon Bedrock (Claude) con contexto específico
- **Ofrece modo interactivo** para consultas en tiempo real

## 🚀 Funcionalidades

### ✅ Procesamiento de Documentos
- Descarga automática de documentos desde S3
- División inteligente en chunks por oraciones
- Indexación para búsqueda rápida

### ✅ Sistema RAG
- Búsqueda semántica por palabras clave
- Selección de chunks más relevantes
- Generación de respuestas contextualizadas

### ✅ Interfaz Flexible
- **Modo línea de comandos**: Para procesamiento batch
- **Modo interactivo**: Para consultas en tiempo real
- **Comandos integrados**: Para gestión de documentos

## 📦 Requisitos Previos

- **Go 1.19+** instalado
- **Cuenta de AWS** activa con credenciales configuradas
- **Bucket de S3** con documentos para procesar
- **Acceso a Amazon Bedrock** habilitado (modelo Claude)

## ⚙️ Instalación

1. **Clonar y configurar:**
   ```bash
   cd 03/
   go mod tidy
   ```

2. **Configurar credenciales AWS:**
   ```bash
   aws configure
   ```

## 🔑 Permisos Necesarios

Tu usuario de IAM necesita:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::tu-bucket-name",
        "arn:aws:s3:::tu-bucket-name/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:aws:bedrock:*::foundation-model/anthropic.claude-v2"
      ]
    }
  ]
}
```

## 🏃‍♂️ Uso

### Modo Línea de Comandos

#### Procesar documento y hacer consulta:
```bash
go run main.go -bucket mi-bucket -s3key documentos/mi-archivo.txt -query "¿Cuál es el tema principal del documento?"
```

#### Solo procesar documento:
```bash
go run main.go -bucket mi-bucket -s3key documentos/mi-archivo.txt
```

#### Solo hacer consulta (documento ya procesado):
```bash
go run main.go -bucket mi-bucket -query "¿Qué dice sobre AWS?"
```

### Modo Interactivo

```bash
go run main.go -bucket mi-bucket
```

Comandos disponibles en modo interactivo:
- `/load <s3-key>` - Cargar documento desde S3
- `/status` - Ver estado del sistema
- `/quit` - Salir
- O escribe directamente tu pregunta

### Ejemplo de Sesión Interactiva

```
🚀 Modo interactivo RAG - AWS Bedrock + S3
Comandos disponibles:
  /load <s3-key>  - Cargar documento desde S3
  /status         - Ver estado del sistema
  /quit           - Salir
O simplemente escribe tu pregunta sobre el documento cargado.

RAG> /load documentos/manual-aws.txt
📄 Cargando documento: documentos/manual-aws.txt
📥 Document downloaded from S3: 15420 bytes
📊 Document split into 23 chunks
✅ Documento cargado exitosamente!

RAG> ¿Qué es Amazon S3?
🤖 Procesando pregunta...

📝 Respuesta:
Amazon S3 (Simple Storage Service) es un servicio de almacenamiento de objetos que ofrece escalabilidad, disponibilidad de datos, seguridad y rendimiento líderes en la industria...

RAG> /quit
👋 ¡Hasta luego!
```

## 🔧 Parámetros de Configuración

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `-region` | Región de AWS | `us-east-1` |
| `-bucket` | Nombre del bucket S3 | **Requerido** |
| `-s3key` | Clave/path del documento en S3 | Opcional |
| `-query` | Pregunta sobre el documento | Opcional |

## 🛠️ Arquitectura del Sistema

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│                 │    │                 │    │                 │
│   Amazon S3     │───▶│   RAG System    │───▶│  Amazon Bedrock │
│   (Documentos)  │    │   (Go App)      │    │   (Claude)      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │                 │
                       │   Usuario       │
                       │   (Terminal)    │
                       │                 │
                       └─────────────────┘
```

### Flujo de Procesamiento

1. **Carga de Documento**: Descarga desde S3
2. **Procesamiento**: División en chunks de ~1000 caracteres
3. **Consulta**: Búsqueda de chunks relevantes
4. **Generación**: Envío a Bedrock con contexto
5. **Respuesta**: Retorno de respuesta generada

## 🔍 Algoritmo de Búsqueda

El sistema utiliza un algoritmo de búsqueda simple pero efectivo:

1. **Tokenización**: Divide la pregunta en palabras
2. **Puntuación**: Cuenta ocurrencias de palabras en cada chunk
3. **Ranking**: Ordena chunks por relevancia
4. **Selección**: Toma los 3 chunks más relevantes
5. **Fallback**: Si no hay coincidencias, usa los primeros chunks

## 🚨 Solución de Problemas

### Error: "bucket name is required"
**Solución**: Especifica el bucket con `-bucket nombre-bucket`

### Error: "no document has been processed yet"
**Solución**: Carga un documento primero con `-s3key` o `/load`

### Error: "error downloading from S3"
**Causas posibles**:
- Archivo no existe en S3
- Permisos insuficientes
- Bucket incorrecto

### Error: "error invoking Claude"
**Causas posibles**:
- Modelo no habilitado en Bedrock
- Permisos insuficientes
- Región incorrecta

## 📊 Optimizaciones Futuras

- [ ] Búsqueda semántica con embeddings
- [ ] Soporte para múltiples formatos (PDF, Word, etc.)
- [ ] Cache de chunks procesados
- [ ] Métricas de relevancia mejoradas
- [ ] Interfaz web
- [ ] Soporte para múltiples modelos

## 🤝 Contribuciones

Las contribuciones son bienvenidas. Para contribuir:

1. Fork el proyecto
2. Crea una rama feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -m 'feat: agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT.

## 🆘 Soporte

Para soporte:
1. Revisa la documentación
2. Verifica la configuración de AWS
3. Consulta los logs de error
4. Abre un issue en el repositorio
