#!/bin/bash

# Script de despliegue para Frontend en AWS S3
# Autor: Sistema EXT3 Simulator
# Fecha: Junio 2025

set -e

echo "🚀 Iniciando despliegue del Frontend en AWS S3..."

# Variables de configuración
BUCKET_NAME="frontend-mia-202307705"
REGION="us-east-1"
DIST_DIR="client/dist"

# Verificar que AWS CLI esté configurado
if ! command -v aws &> /dev/null; then
    echo "❌ Error: AWS CLI no está instalado"
    exit 1
fi

# Verificar credenciales AWS
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ Error: AWS CLI no está configurado correctamente"
    echo "Ejecuta: aws configure"
    exit 1
fi

# Construir el proyecto Frontend
echo "📦 Construyendo el proyecto React..."
cd client
npm install
npm run build
cd ..

# Verificar que el directorio dist existe
if [ ! -d "$DIST_DIR" ]; then
    echo "❌ Error: Directorio de build no encontrado: $DIST_DIR"
    exit 1
fi

# Crear bucket S3 (ignorar si ya existe)
echo "🪣 Creando bucket S3: $BUCKET_NAME"
aws s3 mb s3://$BUCKET_NAME --region $REGION || echo "Bucket ya existe"

# Configurar el bucket para static website hosting
echo "🌐 Configurando Static Website Hosting..."
aws s3api put-bucket-website \
    --bucket $BUCKET_NAME \
    --website-configuration file://website-config.json

# Aplicar política del bucket
echo "🔒 Aplicando política del bucket..."
aws s3api put-bucket-policy \
    --bucket $BUCKET_NAME \
    --policy file://bucket-policy.json

# Subir archivos al bucket
echo "📤 Subiendo archivos al bucket..."
aws s3 sync $DIST_DIR/ s3://$BUCKET_NAME --delete

# Configurar CORS si es necesario
echo "🔄 Configurando CORS..."
aws s3api put-bucket-cors \
    --bucket $BUCKET_NAME \
    --cors-configuration '{
        "CORSRules": [
            {
                "AllowedOrigins": ["*"],
                "AllowedMethods": ["GET", "POST", "PUT", "DELETE"],
                "AllowedHeaders": ["*"],
                "MaxAgeSeconds": 3000
            }
        ]
    }'

# Obtener la URL del sitio web
WEBSITE_URL="http://frontend-mia-202307705.s3-website-us-east-1.amazonaws.com/"

echo "✅ Despliegue completado exitosamente!"
echo "🌍 URL del sitio web: http://frontend-mia-202307705.s3-website-us-east-1.amazonaws.com/"
echo ""
echo "📋 Próximos pasos:"
echo "   1. Verifica que el sitio web funcione correctamente"
echo "   2. Configura el backend en EC2"
echo "   3. Actualiza la URL del API en el frontend si es necesario"
