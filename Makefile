# Makefile for sonar-api-proxy (CentOS 7)

BINARY_NAME=sonar-api-proxy
INSTALL_DIR=/opt/sonar-api-proxy
SYSTEMD_DIR=/etc/systemd/system
SERVICE_NAME=sonar-api-proxy.service
BUILD_DIR=build
GO=go

# CentOS 7 specific
SERVICE_USER=nobody
SERVICE_GROUP=nobody

.PHONY: all build clean install uninstall service start stop restart status enable disable logs

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/main.go

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make install' as root or with sudo"; exit 1; fi
	@mkdir -p $(INSTALL_DIR)/bin
	@mkdir -p $(INSTALL_DIR)/etc
	@mkdir -p $(INSTALL_DIR)/logs
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/bin/
	@chmod 755 $(INSTALL_DIR)/bin/$(BINARY_NAME)
	@chown -R $(SERVICE_USER):$(SERVICE_GROUP) $(INSTALL_DIR)
	@if [ -f "example.env" ]; then \
		cp example.env $(INSTALL_DIR)/etc/sonar-api-proxy.env.example; \
	fi
	@if [ ! -f "$(INSTALL_DIR)/etc/sonar-api-proxy.env" ]; then \
		echo "Creating configuration file..."; \
		cp example.env $(INSTALL_DIR)/etc/sonar-api-proxy.env; \
		chmod 600 $(INSTALL_DIR)/etc/sonar-api-proxy.env; \
		chown $(SERVICE_USER):$(SERVICE_GROUP) $(INSTALL_DIR)/etc/sonar-api-proxy.env; \
		echo "Please edit $(INSTALL_DIR)/etc/sonar-api-proxy.env and add your SONAR_API_TOKEN"; \
	fi
	@echo "Creating systemd service..."
	@echo "[Unit]" > $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "Description=Sonar API Proxy Service" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "After=network.target network-online.target" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "Wants=network-online.target" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "[Service]" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "Type=simple" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "User=$(SERVICE_USER)" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "Group=$(SERVICE_GROUP)" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "WorkingDirectory=$(INSTALL_DIR)" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "EnvironmentFile=$(INSTALL_DIR)/etc/sonar-api-proxy.env" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "ExecStart=$(INSTALL_DIR)/bin/$(BINARY_NAME)" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "Restart=always" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "RestartSec=5" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "StandardOutput=journal" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "StandardError=journal" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "SyslogIdentifier=$(BINARY_NAME)" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "NoNewPrivileges=yes" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "PrivateTmp=yes" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "PrivateDevices=yes" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "ProtectSystem=full" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "ProtectHome=yes" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "LimitNOFILE=4096" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "[Install]" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@echo "WantedBy=multi-user.target" >> $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@chmod 644 $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@systemctl daemon-reload
	@echo "Installation complete!"
	@echo "To start the service, run: make start"
	@echo "To enable at boot, run: make enable"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make uninstall' as root or with sudo"; exit 1; fi
	@if systemctl is-active --quiet $(SERVICE_NAME) 2>/dev/null; then \
		systemctl stop $(SERVICE_NAME); \
	fi
	@if systemctl is-enabled --quiet $(SERVICE_NAME) 2>/dev/null; then \
		systemctl disable $(SERVICE_NAME); \
	fi
	@rm -f $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@rm -f /etc/logrotate.d/$(BINARY_NAME)
	@rm -rf $(INSTALL_DIR)
	@systemctl daemon-reload
	@echo "Uninstallation complete!"

enable:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make enable' as root or with sudo"; exit 1; fi
	systemctl enable $(SERVICE_NAME)
	@echo "Service enabled to start at boot"

disable:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make disable' as root or with sudo"; exit 1; fi
	systemctl disable $(SERVICE_NAME)
	@echo "Service disabled from starting at boot"

start:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make start' as root or with sudo"; exit 1; fi
	systemctl start $(SERVICE_NAME)
	@echo "Service started"

stop:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make stop' as root or with sudo"; exit 1; fi
	systemctl stop $(SERVICE_NAME)
	@echo "Service stopped"

restart:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make restart' as root or with sudo"; exit 1; fi
	systemctl restart $(SERVICE_NAME)
	@echo "Service restarted"

status:
	systemctl status $(SERVICE_NAME)

logs:
	journalctl -u $(SERVICE_NAME) -f

check-selinux:
	@if command -v getenforce >/dev/null 2>&1; then \
		echo "SELinux status: $$(getenforce)"; \
		if [ "$$(getenforce)" = "Enforcing" ]; then \
			echo "Warning: SELinux is enforcing. You may need to adjust contexts if the service fails to start."; \
			echo "Try: make fix-selinux"; \
		fi; \
	else \
		echo "SELinux not found"; \
	fi

fix-selinux:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make fix-selinux' as root or with sudo"; exit 1; fi
	@if command -v chcon >/dev/null 2>&1; then \
		chcon -R -t bin_t $(INSTALL_DIR)/bin/$(BINARY_NAME); \
		echo "SELinux contexts updated"; \
	else \
		echo "SELinux commands not found"; \
	fi

create-logrotate:
	@if [ "$$(id -u)" != "0" ]; then echo "Please run 'make create-logrotate' as root or with sudo"; exit 1; fi
	@echo "$(INSTALL_DIR)/logs/*.log {" > /etc/logrotate.d/$(BINARY_NAME)
	@echo "    daily" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    rotate 7" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    compress" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    delaycompress" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    missingok" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    notifempty" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "    create 0640 $(SERVICE_USER) $(SERVICE_GROUP)" >> /etc/logrotate.d/$(BINARY_NAME)
	@echo "}" >> /etc/logrotate.d/$(BINARY_NAME)
