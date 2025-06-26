#!/bin/bash

# Script de despliegue para Backend en AWS EC2


set -e

echo "🚀 Iniciando configuración del Backend en EC2..."

# Variables de configuración
INSTANCE_ID="i-03b559b62623d97ec"  
KEY_PATH="mia-key"    
EC2_USER="ubuntu"
EC2_HOST="172.31.92.17"    

# Verificar variables
if [ -z "$INSTANCE_ID" ] || [ -z "$KEY_PATH" ] || [ -z "$EC2_HOST" ]; then
    echo "❌ Error: Debes completar las variables de configuración en el script"
    echo "   - INSTANCE_ID: ID de la instancia EC2"
    echo "   - KEY_PATH: Ruta al archivo .pem"
    echo "   - EC2_HOST: IP pública de la instancia"
    exit 1
fi

# Función para ejecutar comandos en EC2
run_remote() {
    ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no "$EC2_USER@$EC2_HOST" "$1"
}

# Función para copiar archivos a EC2
copy_to_ec2() {
    scp -i "$KEY_PATH" -o StrictHostKeyChecking=no -r "$1" "$EC2_USER@$EC2_HOST:$2"
}

echo "📋 Verificando conectividad con la instancia EC2..."
if ! run_remote "echo 'Conexión exitosa'"; then
    echo "❌ Error: No se puede conectar a la instancia EC2"
    exit 1
fi

echo "🔄 Actualizando el sistema..."
run_remote "sudo apt update && sudo apt upgrade -y"

echo "📦 Instalando dependencias..."
run_remote "sudo apt install -y wget curl git build-essential"

echo "🐹 Instalando Go..."
run_remote "
    if ! command -v go &> /dev/null; then
        wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc
        rm go1.21.5.linux-amd64.tar.gz
    fi
"

echo "📁 Creando directorio de trabajo..."
run_remote "mkdir -p ~/ext3-simulator"

echo "📤 Copiando código fuente..."
copy_to_ec2 "server/" "~/ext3-simulator/"

echo "🏗️ Construyendo la aplicación..."
run_remote "
    cd ~/ext3-simulator/server
    export PATH=\$PATH:/usr/local/go/bin
    go mod tidy
    go build -o main .
"

echo "🔧 Configurando servicio systemd..."
run_remote "
    sudo tee /etc/systemd/system/ext3-simulator.service > /dev/null <<EOF
[Unit]
Description=EXT3 Simulator Backend
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/ext3-simulator/server
ExecStart=/home/ubuntu/ext3-simulator/server/main server 8080
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
"

echo "🚀 Iniciando el servicio..."
run_remote "
    sudo systemctl daemon-reload
    sudo systemctl enable ext3-simulator.service
    sudo systemctl start ext3-simulator.service
"

echo "🔍 Verificando el estado del servicio..."
if run_remote "sudo systemctl is-active --quiet ext3-simulator.service"; then
    echo "✅ Servicio iniciado correctamente"
else
    echo "❌ Error: El servicio no se inició correctamente"
    echo "📋 Logs del servicio:"
    run_remote "sudo journalctl -u ext3-simulator.service --lines=20"
    exit 1
fi

echo "🔥 Configurando firewall..."
run_remote "
    sudo ufw allow 22/tcp
    sudo ufw allow 8080/tcp
    sudo ufw --force enable
"

echo "✅ Despliegue completado exitosamente!"
echo "🌍 URL del API: http://$EC2_HOST:8080"
echo ""
echo "📋 Comandos útiles:"
echo "   Ver logs: ssh -i $KEY_PATH $EC2_USER@$EC2_HOST 'sudo journalctl -u ext3-simulator.service -f'"
echo "   Reiniciar: ssh -i $KEY_PATH $EC2_USER@$EC2_HOST 'sudo systemctl restart ext3-simulator.service'"
echo "   Estado: ssh -i $KEY_PATH $EC2_USER@$EC2_HOST 'sudo systemctl status ext3-simulator.service'"
