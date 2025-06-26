# Manual Técnico - Sistema de Archivos EXT3 Simulado

## Tabla de Contenidos
1. [Introducción](#introducción)
2. [Descripción de la Arquitectura del Sistema](#descripción-de-la-arquitectura-del-sistema)
3. [Arquitectura de Despliegue en AWS](#arquitectura-de-despliegue-en-aws)
4. [Explicación de las Estructuras de Datos](#explicación-de-las-estructuras-de-datos)
5. [Módulos Frontend](#módulos-frontend)
6. [Módulos Backend](#módulos-backend)
7. [Protocolos de Comunicación](#protocolos-de-comunicación)
8. [Configuración y Despliegue](#configuración-y-despliegue)

---

## Introducción

Este manual técnico documenta el sistema de archivos EXT3 simulado implementado como una aplicación web full-stack. El sistema proporciona una interfaz gráfica para la administración de discos virtuales, particiones y sistema de archivos, replicando el comportamiento del sistema de archivos EXT3 de Linux.

### Tecnologías Utilizadas
- **Frontend**: React.js con Vite, React Router DOM
- **Backend**: Go (Golang) con servidor HTTP nativo
- **Despliegue**: AWS S3 (Frontend) + AWS EC2 (Backend)
- **Formato de Datos**: Archivos binarios .dsk

---

## Descripción de la Arquitectura del Sistema

### Arquitectura General
![alt text](/image/i1.png)
```

                            │
                            ▼
```
![alt text](/image/i2.png)

### Comunicación entre Módulos

#### Frontend → Backend
1. **Autenticación**: El usuario inicia sesión a través del componente Login
2. **Comandos**: Los comandos se envían desde el componente Console al backend
3. **Respuestas**: El backend procesa y retorna resultados formateados

#### Backend → Sistema de Archivos
1. **Parser**: Analiza comandos entrantes
2. **Executor**: Ejecuta operaciones sobre archivos binarios
3. **Storage**: Maneja la persistencia de datos en archivos .dsk

---

## Arquitectura de Despliegue en AWS

### Diagrama de Despliegue

![alt text](/image/i3.png)
### Componentes de Despliegue

#### 1. Frontend - AWS S3 Static Website Hosting
```
Configuración del Bucket S3:
├── Bucket Name: ext3-simulator-frontend
├── Region: us-east-1
├── Public Access: Enabled
├── Static Website Hosting: Enabled
├── Index Document: index.html
├── Error Document: error.html
└── CORS Configuration: Habilitado para API calls
```

#### 2. Backend - AWS EC2 Instance
```
Configuración de la Instancia EC2:
├── Instance Type: t2.micro (Free Tier)
├── Operating System: Ubuntu 22.04 LTS
├── Storage: 8GB SSD
├── Security Group:
│   ├── Port 22 (SSH): 0.0.0.0/0
│   ├── Port 8080 (HTTP): 0.0.0.0/0
│   └── Port 443 (HTTPS): 0.0.0.0/0 (Opcional)
├── Key Pair: ext3-simulator-key.pem
└── User Data Script: Instalación automática de Go
```

### Flujo de Despliegue

#### Frontend (S3)
1. **Build**: `npm run build` genera archivos estáticos
2. **Upload**: Archivos se suben al bucket S3
3. **Configuration**: Se configura Static Website Hosting
4. **URL**: `http://bucket-name.s3-website-region.amazonaws.com`

#### Backend (EC2)
1. **Instance Setup**: Lanzamiento de instancia Ubuntu
2. **Dependencies**: Instalación de Go 1.21+
3. **Application**: Transferencia del código fuente
4. **Build**: Compilación del binario Go
5. **Service**: Configuración como servicio systemd
6. **Monitoring**: Logs y monitoreo de estado

---

## Explicación de las Estructuras de Datos

### 1. Master Boot Record (MBR)

El MBR es la estructura principal que define la información básica del disco y sus particiones.

```go
type MBR struct {
    Mbr_size           int32      // Tamaño del disco en bytes
    Mbr_creation_date  float32    // Fecha de creación (timestamp)
    Mbr_disk_signature int32      // Firma única del disco
    Mbr_disk_fit       [1]byte    // Algoritmo de ajuste (FF, BF, WF)
    Mbr_partitions     [4]PARTITION // Array de 4 particiones
}
```

**Función**: Define la estructura del disco y contiene información sobre las particiones primarias y extendidas.

**Organización en el archivo .dsk**:
![alt text](/image/i4.png)
### 2. Partition (Partición)

```go
type PARTITION struct {
    Part_status [1]byte    // Estado: 0=inactiva, 1=activa
    Part_type   [1]byte    // Tipo: P=primaria, E=extendida, L=lógica
    Part_fit    [1]byte    // Ajuste: FF, BF, WF
    Part_start  int32      // Byte donde inicia la partición
    Part_size   int32      // Tamaño en bytes
    Part_name   [16]byte   // Nombre de la partición
}
```

### 3. SuperBlock (Superbloque)

El superbloque contiene metadatos críticos del sistema de archivos EXT3.

```go
type SuperBlock struct {
    S_filesystem_type   int32    // Tipo de sistema de archivos
    S_inodes_count      int32    // Cantidad total de inodos
    S_blocks_count      int32    // Cantidad total de bloques
    S_free_inodes_count int32    // Inodos libres
    S_free_blocks_count int32    // Bloques libres
    S_mtime             float32  // Última vez montado
    S_umtime            float32  // Última vez desmontado
    S_mnt_count         int32    // Veces que se ha montado
    S_magic             int32    // Número mágico del sistema
    S_inode_size        int32    // Tamaño de un inodo
    S_block_size        int32    // Tamaño de un bloque
    S_first_ino         int32    // Primer inodo disponible
    S_first_blo         int32    // Primer bloque disponible
    S_bm_inode_start    int32    // Inicio del bitmap de inodos
    S_bm_block_start    int32    // Inicio del bitmap de bloques
    S_inode_start       int32    // Inicio de la tabla de inodos
    S_block_start       int32    // Inicio de los bloques de datos
}
```

**Organización en la partición**:
![alt text](/image/i5.png)

### 4. Inode (Inodo)

Los inodos almacenan metadatos de archivos y directorios.

```go
type Inode struct {
    I_uid   int32      // ID del usuario propietario
    I_gid   int32      // ID del grupo propietario
    I_size  int32      // Tamaño del archivo en bytes
    I_atime float32    // Último acceso
    I_ctime float32    // Creación
    I_mtime float32    // Última modificación
    I_block [15]int32  // Punteros a bloques
    I_type  [1]byte    // Tipo: 0=archivo, 1=carpeta
    I_perm  [3]byte    // Permisos: usuario, grupo, otros
}
```

**Estructura de Punteros**:
```
I_block[0-11]:  Punteros directos a bloques de datos
I_block[12]:    Puntero indirecto simple
I_block[13]:    Puntero indirecto doble
I_block[14]:    Puntero indirecto triple
```

### 5. Bloques de Datos

#### Bloque de Carpeta (Folder Block)
```go
type FolderBlock struct {
    B_content [4]FolderContent  // 4 entradas por bloque
}

type FolderContent struct {
    B_name  [12]byte  // Nombre del archivo/carpeta
    B_inodo int32     // Número de inodo
}
```

#### Bloque de Archivo (File Block)
```go
type FileBlock struct {
    B_content [64]byte  // Contenido del archivo
}
```

#### Bloque de Punteros (Pointer Block)
```go
type PointerBlock struct {
    B_pointers [16]int32  // 16 punteros a otros bloques
}
```

### 6. Journaling (EXT3)

El journaling en EXT3 registra las operaciones antes de ejecutarlas.

```go
type Journal struct {
    Journal_operation [10]byte  // Tipo de operación
    Journal_path      [100]byte // Ruta del archivo
    Journal_content   [100]byte // Contenido de la operación
    Journal_date      float32   // Fecha de la operación
}
```

### Gestión de Archivos .dsk

Los archivos .dsk son archivos binarios que contienen toda la información del disco virtual:

1. **Estructura Física**:
   - Comienzan con el MBR en el byte 0
   - Seguido por las particiones en orden
   - Cada partición puede contener un sistema de archivos EXT3

2. **Operaciones de I/O**:
   - Lectura/escritura binaria directa
   - Uso de offsets calculados para posicionamiento
   - Serialización/deserialización de estructuras Go

3. **Gestión de Espacio**:
   - Bitmaps para rastrear inodos y bloques libres
   - Algoritmos de ajuste: First Fit, Best Fit, Worst Fit
   - Fragmentación controlada mediante bloques de tamaño fijo

---

## Módulos Frontend

### Estructura de Componentes

```
src/
├── components/
│   ├── AnimatedBackground.jsx    // Fondo animado
│   ├── Console.jsx              // Terminal interactiva
│   ├── FileSystemViewer.jsx     // Visualizador de archivos
│   ├── Login.jsx                // Autenticación
│   └── MusicPlayer.jsx          // Reproductor de música
├── context/
│   ├── AuthContext.jsx          // Estado de autenticación
│   └── MusicContext.jsx         // Estado del reproductor
├── routes/
│   ├── AppRoutes.jsx            // Enrutamiento principal
│   └── ProtectedRoute.jsx       // Rutas protegidas
└── services/
    └── api.js                   // Cliente API
```

### Flujo de Datos Frontend

1. **Autenticación**: AuthContext maneja el estado de login
2. **Navegación**: React Router controla las rutas
3. **Comunicación**: API service maneja llamadas HTTP
4. **Estado Global**: Context APIs para datos compartidos

---

## Módulos Backend

### Estructura del Servidor

```
server/
├── analyzer/
│   ├── analyzer.go              // Parser de comandos
│   └── execute.go               // Ejecutor de comandos
├── api/
│   └── server.go                // Servidor HTTP
├── commands/                    // Implementación de comandos
│   ├── mkdisk.go               // Crear disco
│   ├── fdisk.go                // Particionado
│   ├── mkfs.go                 // Formatear
│   ├── mount.go                // Montar partición
│   ├── login.go                // Autenticación
│   ├── mkfile.go               // Crear archivo
│   ├── mkdir.go                // Crear directorio
│   └── [otros comandos]
├── structures/                  // Estructuras de datos
│   ├── mbr.go
│   ├── superblock.go
│   ├── inode.go
│   └── [otras estructuras]
├── stores/
│   └── disk_store.go           // Gestión de discos
└── utils/
    └── utils.go                // Utilidades generales
```

### Flujo de Procesamiento

1. **Recepción**: API recibe comando HTTP
2. **Parsing**: Analyzer extrae comando y parámetros
3. **Validación**: Verificación de sintaxis y permisos
4. **Ejecución**: Command handlers procesan la operación
5. **Persistencia**: Escritura a archivos .dsk
6. **Respuesta**: Retorno de resultado formateado

---

## Protocolos de Comunicación

### API REST Endpoints

```
POST /api/command
Headers:
  Content-Type: application/json
  Authorization: Bearer <token>

Body:
{
  "command": "mkdisk -size=100 -unit=M -path=/tmp/disk1.dsk"
}

Response:
{
  "success": true,
  "message": "Disco creado exitosamente",
  "output": "...",
  "timestamp": "2025-06-25T10:30:00Z"
}
```

### Comandos Soportados

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `mkdisk` | Crear disco virtual | `mkdisk -size=100 -unit=M -path=/tmp/disk1.dsk` |
| `fdisk` | Gestionar particiones | `fdisk -size=50 -unit=M -path=/tmp/disk1.dsk -name=part1` |
| `mkfs` | Formatear partición | `mkfs -id=A1 -type=full` |
| `mount` | Montar partición | `mount -path=/tmp/disk1.dsk -name=part1` |
| `login` | Autenticar usuario | `login -user=root -pass=123 -id=A1` |
| `mkfile` | Crear archivo | `mkfile -path=/home/file.txt -size=100` |
| `mkdir` | Crear directorio | `mkdir -path=/home/folder` |
| `cat` | Mostrar contenido | `cat -path=/home/file.txt` |
| `rep` | Generar reportes | `rep -name=mbr -path=/tmp/disk1.dsk` |

---

## Configuración y Despliegue

### Prerequisitos

#### Frontend
```bash
# Node.js 18+
node --version
npm --version

# Dependencias
npm install
```

#### Backend
```bash
# Go 1.21+
go version

# Módulos Go
go mod tidy
```

### Scripts de Despliegue

#### Frontend (S3)
```bash
#!/bin/bash
# deploy-frontend.sh

# Build del proyecto
npm run build

# Configurar AWS CLI
aws configure set region us-east-1

# Crear bucket S3
aws s3 mb s3://ext3-simulator-frontend

# Configurar política del bucket
aws s3api put-bucket-policy --bucket ext3-simulator-frontend --policy file://bucket-policy.json

# Habilitar static website hosting
aws s3api put-bucket-website --bucket ext3-simulator-frontend --website-configuration file://website-config.json

# Subir archivos
aws s3 sync dist/ s3://ext3-simulator-frontend --delete

echo "Frontend desplegado en: http://ext3-simulator-frontend.s3-website-us-east-1.amazonaws.com"
```

#### Backend (EC2)
```bash
#!/bin/bash
# deploy-backend.sh

# Conectar a la instancia EC2
ssh -i ext3-simulator-key.pem ubuntu@<EC2_IP>

# Instalar Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clonar y construir aplicación
git clone <repository_url>
cd server
go build -o main .

# Crear servicio systemd
sudo tee /etc/systemd/system/ext3-simulator.service > /dev/null <<EOF
[Unit]
Description=EXT3 Simulator Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/server
ExecStart=/home/ubuntu/server/main server 8080
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Habilitar y iniciar servicio
sudo systemctl enable ext3-simulator.service
sudo systemctl start ext3-simulator.service
sudo systemctl status ext3-simulator.service
```

### Configuración de CORS

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "POST", "PUT", "DELETE"],
      "AllowedHeaders": ["*"],
      "MaxAgeSeconds": 3000
    }
  ]
}
```

### Monitoreo y Logs

```bash
# Ver logs del servicio
sudo journalctl -u ext3-simulator.service -f

# Monitorear recursos
htop
df -h
free -m

# Verificar conexiones
netstat -tlnp | grep :8080
```

---

## Conclusión

Este manual técnico proporciona una guía completa para entender, configurar y desplegar el sistema de archivos EXT3 simulado. La arquitectura modular permite escalabilidad y mantenimiento eficiente, mientras que el despliegue en AWS garantiza disponibilidad y rendimiento en la nube.

### Recursos Adicionales

- **Documentación de Go**: https://golang.org/doc/
- **Documentación de React**: https://reactjs.org/docs/
- **AWS S3 Static Hosting**: https://docs.aws.amazon.com/s3/latest/userguide/WebsiteHosting.html
- **AWS EC2 User Guide**: https://docs.aws.amazon.com/ec2/latest/userguide/

---
