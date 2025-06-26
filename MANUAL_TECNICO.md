# Manual Técnico - Sistema de Archivos EXT3 Simulado

## Tabla de Contenidos
1. [Introducción](#introducción)
2. [Descripción de la Arquitectura del Sistema](#descripción-de-la-arquitectura-del-sistema)
3. [Arquitectura de Despliegue en AWS](#arquitectura-de-despliegue-en-aws)
4. [Explicación de las Estructuras de Datos](#explicación-de-las-estructuras-de-datos)
5. [Comandos del Sistema](#comandos-del-sistema)
6. [Módulos Frontend](#módulos-frontend)
7. [Módulos Backend](#módulos-backend)
8. [Protocolos de Comunicación](#protocolos-de-comunicación)
9. [Configuración y Despliegue](#configuración-y-despliegue)

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

El sistema está dividido en tres capas principales:

1. **Capa de Presentación (Frontend React)**
   - Interfaz de usuario web responsive
   - Componentes interactivos para gestión de discos
   - Terminal web para comandos
   - Visualizador de sistema de archivos

2. **Capa de Lógica de Negocio (Backend Go)**
   - API REST para comunicación
   - Parser y analizador de comandos
   - Gestión de estructuras de datos EXT3
   - Sistema de autenticación y permisos

3. **Capa de Persistencia (Archivos .dsk)**
   - Archivos binarios que simulan discos duros
   - Estructuras de datos almacenadas en formato binario
   - Gestión de journaling para EXT3

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

---

## Explicación de las Estructuras de Datos

### 1. Master Boot Record (MBR)

El MBR es la estructura principal que define la información básica del disco y sus particiones.

```go
type MBR struct {
    Mbr_size           int32         // Tamaño del disco en bytes
    Mbr_creation_date  float32       // Fecha de creación (timestamp Unix)
    Mbr_disk_signature int32         // Firma única del disco para identificación
    Mbr_disk_fit       [1]byte       // Algoritmo de ajuste (F=First Fit, B=Best Fit, W=Worst Fit)
    Mbr_partitions     [4]PARTITION  // Array de 4 particiones (máximo 4 particiones primarias/extendidas)
}
```

**Funcionalidades del MBR**:
- Almacena información básica del disco virtual
- Contiene tabla de particiones con máximo 4 entradas
- Gestiona la firma única del disco para identificación
- Define el algoritmo de ajuste para asignación de espacio

**Organización en el archivo .dsk**:
```
Byte 0-3:     Mbr_size (tamaño total del disco)
Byte 4-7:     Mbr_creation_date (timestamp de creación)
Byte 8-11:    Mbr_disk_signature (firma del disco)
Byte 12:      Mbr_disk_fit (algoritmo de ajuste)
Byte 13-...:  Mbr_partitions[4] (información de particiones)
```

### 2. Partition (Partición)

```go
type PARTITION struct {
    Part_status [1]byte    // Estado: '0'=inactiva, '1'=activa/montada
    Part_type   [1]byte    // Tipo: 'P'=primaria, 'E'=extendida, 'L'=lógica
    Part_fit    [1]byte    // Ajuste: 'F'=First Fit, 'B'=Best Fit, 'W'=Worst Fit
    Part_start  int32      // Byte donde inicia la partición en el disco
    Part_size   int32      // Tamaño en bytes de la partición
    Part_name   [16]byte   // Nombre de la partición (string con padding null)
    Part_id     [4]byte    // ID único de la partición cuando está montada
}
```

**Tipos de Particiones**:
- **Primaria (P)**: Partición principal, máximo 4 por disco
- **Extendida (E)**: Contenedor para particiones lógicas
- **Lógica (L)**: Partición dentro de una extendida

### 3. SuperBlock (Superbloque)

El superbloque contiene metadatos críticos del sistema de archivos EXT2/EXT3.

```go
type SuperBlock struct {
    S_filesystem_type   int32    // Tipo: 2=EXT2, 3=EXT3
    S_inodes_count      int32    // Total de inodos creados
    S_blocks_count      int32    // Total de bloques creados
    S_free_inodes_count int32    // Inodos disponibles
    S_free_blocks_count int32    // Bloques disponibles
    S_mtime             float32  // Última vez montado (timestamp)
    S_umtime            float32  // Última vez desmontado (timestamp)
    S_mnt_count         int32    // Cantidad de veces montado
    S_magic             int32    // Número mágico: 0xEF53
    S_inode_size        int32    // Tamaño de un inodo en bytes
    S_block_size        int32    // Tamaño de un bloque en bytes
    S_first_ino         int32    // Posición del primer inodo libre
    S_first_blo         int32    // Posición del primer bloque libre
    S_bm_inode_start    int32    // Inicio del bitmap de inodos
    S_bm_block_start    int32    // Inicio del bitmap de bloques
    S_inode_start       int32    // Inicio de la tabla de inodos
    S_block_start       int32    // Inicio de los bloques de datos
}
```

**Organización en la partición EXT3**:
```
[SuperBlock][Journal Area][Bitmap Inodos][Bitmap Bloques][Tabla Inodos][Bloques de Datos]
```

**Organización en la partición EXT2**:
```
[SuperBlock][Bitmap Inodos][Bitmap Bloques][Tabla Inodos][Bloques de Datos]
```

### 4. Inode (Inodo)

Los inodos almacenan metadatos completos de archivos y directorios.

```go
type Inode struct {
    I_uid   int32      // ID del usuario propietario
    I_gid   int32      // ID del grupo propietario
    I_size  int32      // Tamaño del archivo/directorio en bytes
    I_atime float32    // Último tiempo de acceso (timestamp)
    I_ctime float32    // Tiempo de creación (timestamp)
    I_mtime float32    // Última modificación (timestamp)
    I_block [15]int32  // Array de punteros a bloques
    I_type  [1]byte    // Tipo: '0'=directorio, '1'=archivo
    I_perm  [3]byte    // Permisos en octal: [propietario][grupo][otros]
}
```

**Estructura de Punteros en I_block**:
```
I_block[0-11]:  Punteros directos (12 bloques directos)
I_block[12]:    Puntero indirecto simple (apunta a bloque con 16 punteros)
I_block[13]:    Puntero indirecto doble (apunta a bloque con punteros a bloques de punteros)
I_block[14]:    Puntero indirecto triple (tres niveles de indirección)
```

**Sistema de Permisos**:
```
Valores octales para permisos:
0: --- (sin permisos)
1: --x (solo ejecución)
2: -w- (solo escritura)
3: -wx (escritura y ejecución)
4: r-- (solo lectura)
5: r-x (lectura y ejecución)
6: rw- (lectura y escritura)
7: rwx (todos los permisos)
```

### 5. Bloques de Datos

#### Bloque de Carpeta (FolderBlock)
```go
type FolderBlock struct {
    B_content [4]FolderContent  // Máximo 4 entradas por bloque
}

type FolderContent struct {
    B_name  [12]byte  // Nombre del archivo/carpeta (máximo 12 caracteres)
    B_inodo int32     // Número de inodo al que apunta (-1 si está vacío)
}
```

**Entradas Especiales**:
- `"."`: Referencia al directorio actual
- `".."`: Referencia al directorio padre

#### Bloque de Archivo (FileBlock)
```go
type FileBlock struct {
    B_content [64]byte  // Contenido del archivo (64 bytes por bloque)
}
```

#### Bloque de Punteros (PointerBlock)
```go
type PointerBlock struct {
    B_pointers [16]int32  // Array de 16 punteros a otros bloques
}
```

### 6. Journaling (EXT3)

El journaling registra operaciones antes de ejecutarlas para garantizar consistencia.

```go
type Journal struct {
    J_next    int32          // Puntero al siguiente journal (-1 si es el último)
    J_content Information    // Información de la operación
}

type Information struct {
    I_operation [10]byte   // Tipo de operación (mkfile, mkdir, login, etc.)
    I_path      [74]byte   // Ruta del archivo/directorio afectado
    I_content   [64]byte   // Contenido adicional de la operación
    I_date      float32    // Timestamp de la operación
}
```

**Operaciones Registradas**:
- `mkfile`: Creación de archivos
- `mkdir`: Creación de directorios
- `login`: Inicio de sesión de usuarios
- `mkgrp`: Creación de grupos
- `mkusr`: Creación de usuarios
- `rmgrp`: Eliminación de grupos
- `rmusr`: Eliminación de usuarios

### 7. Gestión de Usuarios y Grupos

El sistema utiliza un archivo especial `/users.txt` para gestionar usuarios y grupos:

```
Formato del archivo users.txt:
[ID],[Tipo],[Grupo/Usuario],[Password]

Ejemplos:
1,G,root                    # Grupo root con ID 1
1,U,root,root,123          # Usuario root, grupo root, password 123
2,G,users                  # Grupo users con ID 2
3,U,user1,users,pass123    # Usuario user1, grupo users
```

### 8. Estructuras de Control del Sistema

#### Variables Globales del Store
```go
var (
    MountedPartitions map[string]string // ID_particion -> ruta_disco
    LogedIdPartition  string            // ID de la partición actual
    LogedUser         string            // Usuario logueado actual
    LoadedDiskPaths   map[string]string // Letra_disco -> ruta_archivo
)
```

---

## Comandos del Sistema

### 1. Gestión de Discos

#### MKDISK - Crear Disco Virtual
```bash
mkdisk -size=<tamaño> -unit=<unidad> -path=<ruta>
```

**Parámetros**:
- `-size`: Tamaño del disco (requerido)
- `-unit`: Unidad de medida (B=bytes, K=kilobytes, M=megabytes)
- `-path`: Ruta donde crear el archivo .dsk (requerido)

**Funcionalidad**:
- Crea un archivo binario .dsk del tamaño especificado
- Inicializa el MBR con valores por defecto
- Establece la fecha de creación y firma del disco
- Valida que la ruta de destino exista

**Ejemplo**:
```bash
mkdisk -size=100 -unit=M -path="/home/user/disco1.dsk"
```

#### RMDISK - Eliminar Disco Virtual
```bash
rmdisk -path=<ruta>
```

**Parámetros**:
- `-path`: Ruta del archivo .dsk a eliminar (requerido)

**Funcionalidad**:
- Elimina físicamente el archivo .dsk del sistema
- Limpia referencias del disco en memoria
- Desmonta particiones asociadas automáticamente

### 2. Gestión de Particiones

#### FDISK - Gestionar Particiones
```bash
fdisk -size=<tamaño> -unit=<unidad> -path=<ruta> -type=<tipo> -fit=<ajuste> -name=<nombre>
```

**Parámetros**:
- `-size`: Tamaño de la partición (requerido)
- `-unit`: Unidad de medida (B, K, M)
- `-path`: Ruta del disco .dsk (requerido)
- `-type`: Tipo de partición (P=primaria, E=extendida, L=lógica)
- `-fit`: Algoritmo de ajuste (FF=First Fit, BF=Best Fit, WF=Worst Fit)
- `-name`: Nombre de la partición (requerido)

**Funcionalidades**:
- Crear particiones primarias (máximo 4)
- Crear particiones extendidas (máximo 1)
- Crear particiones lógicas dentro de extendidas
- Aplicar algoritmos de ajuste para asignación de espacio
- Validar que no haya solapamiento entre particiones

**Algoritmos de Ajuste**:
- **First Fit**: Asigna el primer espacio disponible que sea suficiente
- **Best Fit**: Asigna el espacio más pequeño que sea suficiente
- **Worst Fit**: Asigna el espacio más grande disponible

#### MOUNT - Montar Partición
```bash
mount -driveletter=<letra> -name=<nombre>
```

**Parámetros**:
- `-driveletter`: Letra del disco (A, B, C, etc.) (requerido)
- `-name`: Nombre de la partición a montar (requerido)

**Funcionalidad**:
- Genera un ID único para la partición montada
- Actualiza el estado de la partición a activa
- Registra la partición en el sistema de montaje
- Permite acceso al sistema de archivos de la partición

#### UNMOUNT - Desmontar Partición
```bash
unmount -id=<id_particion>
```

**Parámetros**:
- `-id`: ID de la partición montada (requerido)

**Funcionalidad**:
- Desmonta la partición del sistema
- Actualiza el estado a inactiva
- Cierra sesiones de usuario en la partición
- Limpia referencias en memoria

### 3. Sistema de Archivos

#### MKFS - Formatear Partición
```bash
mkfs -id=<id_particion> -type=<tipo> -fs=<sistema>
```

**Parámetros**:
- `-id`: ID de la partición montada (requerido)
- `-type`: Tipo de formateo (full=completo)
- `-fs`: Sistema de archivos (2fs=EXT2, 3fs=EXT3)

**Funcionalidad**:
- Inicializa el superbloque con metadatos del sistema
- Crea bitmaps de inodos y bloques
- Inicializa la tabla de inodos
- Crea el directorio raíz (/)
- Para EXT3: configura el área de journaling
- Crea el archivo `/users.txt` con usuario root por defecto

**Estructura Inicial**:
```
Inodo 0: Directorio raíz (/)
Inodo 1: Archivo users.txt
Usuario inicial: root/root/123
Grupo inicial: root
```

### 4. Gestión de Usuarios y Grupos

#### LOGIN - Iniciar Sesión
```bash
login -user=<usuario> -pass=<password> -id=<id_particion>
```

**Parámetros**:
- `-user`: Nombre de usuario (requerido)
- `-pass`: Contraseña del usuario (requerido)  
- `-id`: ID de la partición donde autenticar (requerido)

**Funcionalidad**:
- Valida credenciales contra `/users.txt`
- Establece sesión activa del usuario
- Configura permisos según el tipo de usuario
- Registra la operación en el journal (EXT3)

#### LOGOUT - Cerrar Sesión
```bash
logout
```

**Funcionalidad**:
- Cierra la sesión actual del usuario
- Limpia variables de estado de sesión
- Registra la operación en el journal (EXT3)

#### MKGRP - Crear Grupo
```bash
mkgrp -name=<nombre_grupo>
```

**Parámetros**:
- `-name`: Nombre del grupo (requerido)

**Funcionalidad**:
- Requiere usuario root logueado
- Valida que el nombre no exista
- Asigna ID único secuencial
- Actualiza archivo `/users.txt`
- Registra en journal (EXT3)

#### RMGRP - Eliminar Grupo
```bash
rmgrp -name=<nombre_grupo>
```

**Parámetros**:
- `-name`: Nombre del grupo a eliminar (requerido)

**Funcionalidad**:
- Requiere usuario root logueado
- Valida que el grupo exista
- Marca como eliminado (ID=0) en `/users.txt`
- Registra en journal (EXT3)

#### MKUSR - Crear Usuario
```bash
mkusr -user=<usuario> -pass=<password> -grp=<grupo>
```

**Parámetros**:
- `-user`: Nombre de usuario (requerido, máximo 10 caracteres)
- `-pass`: Contraseña (requerido)
- `-grp`: Grupo al que pertenece (requerido)

**Funcionalidad**:
- Requiere usuario root logueado
- Valida que el usuario no exista
- Valida que el grupo exista
- Asigna ID único secuencial
- Actualiza archivo `/users.txt`
- Registra en journal (EXT3)

#### RMUSR - Eliminar Usuario
```bash
rmusr -user=<usuario>
```

**Parámetros**:
- `-user`: Nombre del usuario a eliminar (requerido)

**Funcionalidad**:
- Requiere usuario root logueado
- Valida que el usuario exista
- Marca como eliminado (ID=0) en `/users.txt`
- Registra en journal (EXT3)

### 5. Gestión de Archivos y Directorios

#### MKFILE - Crear Archivo
```bash
mkfile -path=<ruta> -r -size=<tamaño> -cont=<contenido>
```

**Parámetros**:
- `-path`: Ruta completa del archivo (requerido)
- `-r`: Crear directorios padre si no existen (opcional)
- `-size`: Tamaño del archivo en bytes (opcional)
- `-cont`: Contenido del archivo (opcional)

**Funcionalidad**:
- Crea archivo en el sistema de archivos EXT3
- Asigna inodos y bloques según el tamaño
- Establece permisos del usuario actual
- Opcionalmente crea estructura de directorios
- Registra en journal (EXT3)

#### MKDIR - Crear Directorio
```bash
mkdir -path=<ruta> -r
```

**Parámetros**:
- `-path`: Ruta completa del directorio (requerido)
- `-r`: Crear directorios padre si no existen (opcional)

**Funcionalidad**:
- Crea directorio en el sistema de archivos
- Asigna inodo de tipo directorio
- Establece entradas "." y ".."
- Configura permisos del usuario actual
- Registra en journal (EXT3)

#### CAT - Mostrar Contenido
```bash
cat -file1=<ruta1> -file2=<ruta2> ... -fileN=<rutaN>
```

**Parámetros**:
- `-fileN`: Ruta de archivo a mostrar (al menos uno requerido)

**Funcionalidad**:
- Lee contenido de uno o múltiples archivos
- Concatena el contenido si hay múltiples archivos
- Valida permisos de lectura del usuario
- Muestra contenido completo del archivo

#### REMOVE - Eliminar Archivo
```bash
remove -path=<ruta>
```

**Parámetros**:
- `-path`: Ruta del archivo a eliminar (requerido)

**Funcionalidad**:
- Elimina archivo del sistema de archivos
- Libera inodos y bloques asignados
- Actualiza bitmaps de disponibilidad
- Valida permisos de escritura
- Registra en journal (EXT3)

#### EDIT - Editar Archivo
```bash
edit -path=<ruta> -cont=<contenido>
```

**Parámetros**:
- `-path`: Ruta del archivo a editar (requerido)
- `-cont`: Nuevo contenido del archivo (requerido)

**Funcionalidad**:
- Modifica contenido de archivo existente
- Reasigna bloques si cambia el tamaño
- Actualiza timestamp de modificación
- Valida permisos de escritura
- Registra en journal (EXT3)

#### RENAME - Renombrar Archivo/Directorio
```bash
rename -path=<ruta> -name=<nuevo_nombre>
```

**Parámetros**:
- `-path`: Ruta actual del archivo/directorio (requerido)
- `-name`: Nuevo nombre (requerido)

**Funcionalidad**:
- Cambia nombre de archivo o directorio
- Actualiza entrada en directorio padre
- Mantiene inodo y permisos originales
- Valida permisos de escritura en directorio padre
- Registra en journal (EXT3)

#### COPY - Copiar Archivo
```bash
copy -path=<origen> -dest=<destino>
```

**Parámetros**:
- `-path`: Ruta del archivo origen (requerido)
- `-dest`: Ruta de destino (requerido)

**Funcionalidad**:
- Crea copia exacta del archivo
- Asigna nuevos inodos y bloques
- Mantiene contenido pero actualiza metadatos
- Valida permisos de lectura en origen y escritura en destino
- Registra en journal (EXT3)

#### MOVE - Mover Archivo
```bash
move -path=<origen> -dest=<destino>
```

**Parámetros**:
- `-path`: Ruta actual del archivo (requerido)
- `-dest`: Nueva ruta del archivo (requerido)

**Funcionalidad**:
- Mueve archivo de ubicación
- Mantiene el mismo inodo
- Actualiza entradas de directorio
- Valida permisos en origen y destino
- Registra en journal (EXT3)

#### FIND - Buscar Archivos
```bash
find -path=<ruta> -name=<nombre>
```

**Parámetros**:
- `-path`: Directorio donde buscar (requerido)
- `-name`: Nombre o patrón a buscar (requerido)

**Funcionalidad**:
- Busca archivos recursivamente en directorios
- Admite búsqueda por nombre exacto
- Valida permisos de lectura en directorios
- Retorna rutas completas de archivos encontrados

#### CHOWN - Cambiar Propietario
```bash
chown -path=<ruta> -user=<usuario> -r
```

**Parámetros**:
- `-path`: Ruta del archivo/directorio (requerido)
- `-user`: Nuevo usuario propietario (requerido)
- `-r`: Aplicar recursivamente (opcional)

**Funcionalidad**:
- Cambia propietario de archivo o directorio
- Requiere permisos de administrador o ser propietario
- Opcionalmente aplica cambio recursivamente
- Actualiza metadatos del inodo
- Registra en journal (EXT3)

#### CHMOD - Cambiar Permisos
```bash
chmod -path=<ruta> -ugo=<permisos> -r
```

**Parámetros**:
- `-path`: Ruta del archivo/directorio (requerido)
- `-ugo`: Permisos en formato octal (ejemplo: 755) (requerido)
- `-r`: Aplicar recursivamente (opcional)

**Funcionalidad**:
- Modifica permisos de archivo o directorio
- Requiere ser propietario o root
- Formato octal: [propietario][grupo][otros]
- Opcionalmente aplica cambio recursivamente
- Registra en journal (EXT3)

### 6. Reportes del Sistema

#### REP - Generar Reportes
```bash
rep -id=<id_particion> -path=<ruta_salida> -name=<tipo_reporte> -ruta=<ruta_archivo>
```

**Parámetros**:
- `-id`: ID de la partición (requerido)
- `-path`: Ruta donde guardar el reporte (requerido)
- `-name`: Tipo de reporte (requerido)
- `-ruta`: Ruta específica para algunos reportes (opcional)

**Tipos de Reportes Disponibles**:

1. **mbr**: Reporte del Master Boot Record
   - Muestra información del disco y particiones
   - Incluye tabla de particiones con detalles

2. **disk**: Reporte gráfico del uso del disco
   - Visualización gráfica de particiones
   - Porcentajes de uso de espacio

3. **inode**: Reporte de la tabla de inodos
   - Lista todos los inodos con sus metadatos
   - Estado de cada inodo (libre/ocupado)

4. **block**: Reporte de bloques de datos
   - Contenido de bloques de archivos y directorios
   - Estado de cada bloque

5. **bm_inode**: Bitmap de inodos
   - Visualización del bitmap de inodos
   - Estados: 0=libre, 1=ocupado

6. **bm_block**: Bitmap de bloques
   - Visualización del bitmap de bloques
   - Estados: 0=libre, 1=ocupado

7. **tree**: Árbol del sistema de archivos
   - Estructura jerárquica de directorios y archivos
   - Navegación completa desde la raíz

8. **sb**: Reporte del superbloque
   - Metadatos completos del sistema de archivos
   - Estadísticas de uso

9. **file**: Contenido de archivo específico
   - Requiere parámetro `-ruta`
   - Muestra contenido completo del archivo

10. **ls**: Listado de directorio específico
    - Requiere parámetro `-ruta`
    - Equivalente al comando ls de Linux

11. **journaling**: Reporte del journal (solo EXT3)
    - Lista todas las operaciones registradas
    - Timestamps y detalles de cada operación

---

## Módulos Frontend

### Estructura de Componentes

```
src/
├── components/
│   ├── AnimatedBackground.jsx    // Fondo animado con partículas
│   ├── Console.jsx              // Terminal interactiva web
│   ├── FileSystemViewer.jsx     // Explorador de archivos
│   ├── Login.jsx                // Formulario de autenticación
│   └── MusicPlayer.jsx          // Reproductor de música ambiental
├── context/
│   ├── AuthContext.jsx          // Estado global de autenticación
│   └── MusicContext.jsx         // Estado global del reproductor
├── routes/
│   ├── AppRoutes.jsx            // Configuración de rutas principales
│   └── ProtectedRoute.jsx       // Rutas que requieren autenticación
└── services/
    └── api.js                   // Cliente HTTP para comunicación con backend
```

### Flujo de Datos Frontend

1. **Autenticación**: AuthContext maneja el estado de login global
2. **Navegación**: React Router controla las rutas y protección
3. **Comunicación**: API service centraliza llamadas HTTP al backend
4. **Estado Global**: Context APIs para datos compartidos entre componentes

---

## Módulos Backend

### Estructura del Servidor

```
server/
├── analyzer/
│   └── analyzer.go              // Parser y analizador de comandos
├── api/
│   └── server.go                // Servidor HTTP con endpoints REST
├── commands/                    // Implementación de todos los comandos
│   ├── mkdisk.go               // Crear disco virtual
│   ├── rmdisk.go               // Eliminar disco
│   ├── fdisk.go                // Gestión de particiones
│   ├── mount.go                // Montar partición
│   ├── unmount.go              // Desmontar partición
│   ├── mkfs.go                 // Formatear partición
│   ├── login.go                // Autenticación de usuarios
│   ├── logout.go               // Cerrar sesión
│   ├── mkgrp.go                // Crear grupo
│   ├── rmgrp.go                // Eliminar grupo
│   ├── mkusr.go                // Crear usuario
│   ├── rmusr.go                // Eliminar usuario
│   ├── mkfile.go               // Crear archivo
│   ├── mkdir.go                // Crear directorio
│   ├── cat.go                  // Mostrar contenido
│   ├── remove.go               // Eliminar archivo
│   ├── edit.go                 // Editar archivo
│   ├── rename.go               // Renombrar
│   ├── copy.go                 // Copiar archivo
│   ├── move.go                 // Mover archivo
│   ├── find.go                 // Buscar archivos
│   ├── chown.go                // Cambiar propietario
│   ├── chmod.go                // Cambiar permisos
│   └── rep.go                  // Generar reportes
├── structures/                  // Estructuras de datos del sistema
│   ├── mbr.go                  // Master Boot Record
│   ├── partition.go            // Particiones
│   ├── superblock.go           // Superbloque EXT2/EXT3
│   ├── inode.go                // Inodos
│   ├── blocks.go               // Bloques de datos
│   └── journal.go              // Journal para EXT3
├── stores/
│   ├── store.go                // Gestión global del estado
│   └── disk_store.go           // Gestión específica de discos
├── utils/
│   └── utils.go                // Utilidades generales del sistema
├── console/
│   └── console.go              // Utilidades para output de consola
└── reports/
    └── reports.go              // Generación de reportes
```

### Flujo de Procesamiento de Comandos

1. **Recepción HTTP**: API recibe comando a través de endpoint REST
2. **Parsing**: Analyzer extrae comando y valida sintaxis
3. **Validación**: Verificación de parámetros y permisos de usuario
4. **Ejecución**: Command handler específico procesa la operación
5. **Persistencia**: Escritura/lectura de archivos .dsk binarios
6. **Respuesta**: Retorno de resultado formateado en JSON

---

## Protocolos de Comunicación

### API REST Endpoints

#### Comando Individual
```http
POST /api/command
Content-Type: application/json

{
  "command": "mkdisk -size=100 -unit=M -path=/tmp/disk1.dsk"
}

Response:
{
  "success": true,
  "message": "Comando ejecutado exitosamente",
  "data": "MKDISK: disco creado exitosamente",
  "error": null
}
```

#### Comandos en Lote
```http
POST /api/batch
Content-Type: application/json

{
  "commands": [
    "mkdisk -size=100 -unit=M -path=/tmp/disk1.dsk",
    "fdisk -size=50 -unit=M -path=/tmp/disk1.dsk -name=part1 -type=P",
    "mount -driveletter=A -name=part1"
  ]
}

Response:
{
  "success": true,
  "results": [...],
  "summary": {
    "total": 3,
    "success": 3,
    "error": 0
  }
}
```

#### Información del Sistema
```http
GET /api/disks
Response:
{
  "success": true,
  "disks": [
    {
      "id": "A",
      "name": "A", 
      "size": "100.0 MB",
      "status": "Disponible",
      "path": "/tmp/disk1.dsk"
    }
  ]
}

GET /api/partitions?disk=A
Response:
{
  "success": true,
  "partitions": [...],
  "diskId": "A",
  "total": 2
}

GET /api/filesystem?partition=A105&path=/
Response:
{
  "success": true,
  "data": {
    "folders": [...],
    "files": [...]
  },
  "path": "/"
}
```

---

## Configuración y Despliegue

### Prerequisitos del Sistema

#### Frontend
```bash
# Node.js 18+ y npm
node --version  # v18.0.0+
npm --version   # 8.0.0+

# Instalación de dependencias
npm install
```

#### Backend
```bash
# Go 1.21+
go version     # go1.21.0+

# Verificación de módulos
go mod tidy
go mod verify
```

### Scripts de Despliegue Automatizado

#### Frontend en AWS S3
```bash
#!/bin/bash
# deploy-frontend.sh

echo "🚀 Iniciando despliegue del frontend..."

# Construir aplicación React
echo "📦 Construyendo aplicación..."
npm run build

# Configurar AWS CLI si no está configurado
if ! aws sts get-caller-identity > /dev/null 2>&1; then
    echo "⚠️ AWS CLI no configurado. Configurando..."
    aws configure
fi

# Variables de configuración
BUCKET_NAME="ext3-simulator-frontend-$(date +%s)"
REGION="us-east-1"

# Crear bucket S3
echo "🪣 Creando bucket S3: $BUCKET_NAME"
aws s3 mb s3://$BUCKET_NAME --region $REGION

# Configurar política pública del bucket
echo "🔐 Configurando política del bucket..."
cat > bucket-policy.json << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::$BUCKET_NAME/*"
        }
    ]
}
EOF

aws s3api put-bucket-policy --bucket $BUCKET_NAME --policy file://bucket-policy.json

# Configurar static website hosting
echo "🌐 Configurando static website hosting..."
cat > website-config.json << EOF
{
    "IndexDocument": {
        "Suffix": "index.html"
    },
    "ErrorDocument": {
        "Key": "error.html"
    }
}
EOF

aws s3api put-bucket-website --bucket $BUCKET_NAME --website-configuration file://website-config.json

# Subir archivos con configuración optimizada
echo "📤 Subiendo archivos al bucket..."
aws s3 sync dist/ s3://$BUCKET_NAME \
    --delete \
    --cache-control "public, max-age=31536000" \
    --exclude "*.html" \
    --exclude "service-worker.js"

# Subir archivos HTML sin cache
aws s3 sync dist/ s3://$BUCKET_NAME \
    --delete \
    --cache-control "no-cache" \
    --include "*.html" \
    --include "service-worker.js"

# URL final
WEBSITE_URL="http://$BUCKET_NAME.s3-website-$REGION.amazonaws.com"
echo "✅ Frontend desplegado exitosamente!"
echo "🔗 URL: $WEBSITE_URL"

# Limpiar archivos temporales
rm bucket-policy.json website-config.json

echo "🎉 Despliegue completado!"
```

#### Backend en AWS EC2
```bash
#!/bin/bash
# deploy-backend.sh

echo "🚀 Iniciando despliegue del backend..."

# Variables de configuración
EC2_USER="ubuntu"
EC2_HOST="your-ec2-public-ip"
KEY_PATH="./ext3-simulator-key.pem"
APP_NAME="ext3-simulator"

# Verificar que la clave SSH existe
if [ ! -f "$KEY_PATH" ]; then
    echo "❌ Archivo de clave SSH no encontrado: $KEY_PATH"
    exit 1
fi

# Configurar permisos de la clave
chmod 400 $KEY_PATH

echo "📦 Preparando archivos para transferencia..."

# Crear archivo de instalación remota
cat > install-backend.sh << 'EOF'
#!/bin/bash
set -e

echo "🔧 Instalando dependencias..."

# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar Go si no está instalado
if ! command -v go &> /dev/null; then
    echo "⬇️ Descargando e instalando Go..."
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz
    
    # Configurar PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
fi

# Verificar instalación de Go
go version

echo "✅ Dependencias instaladas correctamente"
EOF

# Transferir archivos al servidor
echo "📤 Transfiriendo archivos al servidor..."
scp -i $KEY_PATH -r server/ $EC2_USER@$EC2_HOST:/home/$EC2_USER/
scp -i $KEY_PATH install-backend.sh $EC2_USER@$EC2_HOST:/home/$EC2_USER/

# Ejecutar instalación en el servidor
echo "🔧 Ejecutando instalación en el servidor..."
ssh -i $KEY_PATH $EC2_USER@$EC2_HOST << 'EOF'
# Ejecutar script de instalación
chmod +x install-backend.sh
./install-backend.sh

# Ir al directorio del servidor y construir aplicación
cd server
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o main .

# Crear directorio para archivos de disco
sudo mkdir -p /home/ubuntu/MIA_P2_202307705_1VAC1S2025/test/
sudo chown $USER:$USER /home/ubuntu/MIA_P2_202307705_1VAC1S2025/test/

echo "🚀 Aplicación construida exitosamente"
EOF

# Crear archivo de servicio systemd
echo "⚙️ Configurando servicio systemd..."
cat > ext3-simulator.service << EOF
[Unit]
Description=EXT3 Simulator Backend API
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=$EC2_USER
WorkingDirectory=/home/$EC2_USER/server
ExecStart=/home/$EC2_USER/server/main server 8080
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Transferir y configurar servicio
scp -i $KEY_PATH ext3-simulator.service $EC2_USER@$EC2_HOST:/home/$EC2_USER/
ssh -i $KEY_PATH $EC2_USER@$EC2_HOST << 'EOF'
# Instalar servicio systemd
sudo mv ext3-simulator.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ext3-simulator.service
sudo systemctl start ext3-simulator.service

# Verificar estado del servicio
sleep 3
sudo systemctl status ext3-simulator.service --no-pager
EOF

# Configurar firewall si es necesario
echo "🔥 Configurando firewall..."
ssh -i $KEY_PATH $EC2_USER@$EC2_HOST << 'EOF'
# Configurar UFW si está instalado
if command -v ufw &> /dev/null; then
    sudo ufw allow 8080/tcp
    sudo ufw allow ssh
    echo "🛡️ Firewall configurado"
fi
EOF

echo "✅ Backend desplegado exitosamente!"
echo "🔗 API URL: http://$EC2_HOST:8080"
echo "📊 Estado del servicio:"
ssh -i $KEY_PATH $EC2_USER@$EC2_HOST 'sudo systemctl is-active ext3-simulator.service'

# Limpiar archivos temporales
rm install-backend.sh ext3-simulator.service

echo "🎉 Despliegue completado!"
```

### Monitoreo y Logs

#### Verificar Estado del Backend
```bash
# Conectar al servidor
ssh -i ext3-simulator-key.pem ubuntu@<EC2_IP>

# Ver logs en tiempo real
sudo journalctl -u ext3-simulator.service -f

# Verificar estado del servicio
sudo systemctl status ext3-simulator.service

# Reiniciar servicio si es necesario
sudo systemctl restart ext3-simulator.service

# Ver uso de recursos
htop
df -h
free -m

# Verificar conectividad
curl http://localhost:8080/api/health
```

#### Logs de Aplicación
```bash
# Logs completos del servicio
sudo journalctl -u ext3-simulator.service --since "1 hour ago"

# Logs con filtros
sudo journalctl -u ext3-simulator.service | grep ERROR
sudo journalctl -u ext3-simulator.service | grep "Comando ejecutado"

# Configurar rotación de logs
sudo tee /etc/logrotate.d/ext3-simulator << EOF
/var/log/ext3-simulator/*.log {
    daily
    missingok
    rotate 7
    compress
    notifempty
    create 644 ubuntu ubuntu
}
EOF
```

### Troubleshooting Común

#### Problemas de Conexión
```bash
# Verificar que el puerto esté abierto
netstat -tlnp | grep :8080

# Verificar Security Group en AWS EC2
# - Puerto 8080 debe estar abierto para 0.0.0.0/0

# Verificar conectividad desde cliente
curl -v http://<EC2_IP>:8080/api/health
```

#### Problemas de Permisos
```bash
# Verificar permisos del directorio de archivos
ls -la /home/ubuntu/MIA_P2_202307705_1VAC1S2025/test/

# Corregir permisos si es necesario
sudo chown -R ubuntu:ubuntu /home/ubuntu/MIA_P2_202307705_1VAC1S2025/
chmod 755 /home/ubuntu/MIA_P2_202307705_1VAC1S2025/test/
```

#### Problemas de Memoria/Disco
```bash
# Verificar espacio en disco
df -h

# Verificar memoria
free -m

# Limpiar archivos .dsk antiguos si es necesario
find /home/ubuntu/MIA_P2_202307705_1VAC1S2025/test/ -name "*.dsk" -mtime +7 -delete
```

---

## Conclusión

Este manual técnico proporciona una documentación completa del sistema de archivos EXT3 simulado, incluyendo todas las estructuras de datos, comandos disponibles, arquitectura del sistema y procedimientos de despliegue. El sistema implementa de manera fiel las características del sistema de archivos EXT3 de Linux, proporcionando una plataforma educativa y funcional para el aprendizaje de sistemas operativos y gestión de archivos.

### Características Destacadas

1. **Fidelidad al EXT3 Real**: Implementación precisa de estructuras como inodos, superbloques y journaling
2. **Interfaz Web Moderna**: Frontend React con experiencia de usuario intuitiva
3. **API REST Completa**: Backend Go con endpoints bien documentados
4. **Despliegue en la Nube**: Configuración automática en AWS S3 y EC2
5. **Sistema de Permisos**: Gestión completa de usuarios, grupos y permisos
6. **Journaling Funcional**: Registro de operaciones para consistencia de datos
7. **Reportes Detallados**: Múltiples tipos de reportes para análisis del sistema

### Recursos Adicionales

- **Documentación de Go**: https://golang.org/doc/
- **Documentación de React**: https://reactjs.org/docs/
- **AWS S3 Static Hosting**: https://docs.aws.amazon.com/s3/latest/userguide/WebsiteHosting.html
- **AWS EC2 User Guide**: https://docs.aws.amazon.com/ec2/latest/userguide/
- **Especificación EXT3**: https://ext4.wiki.kernel.org/index.php/Ext3_Design

---
