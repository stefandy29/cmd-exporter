#!/bin/bash
filename=cmd-exporter
useradd --no-create-home --shell /bin/false $filename
chmod +x $filename
cp $filename /usr/local/bin

chown $filename:$filename /usr/local/bin/$filename

mkdir -p /etc/$filename
chown -R $filename:$filename /etc/$filename
chmod 750 -R /etc/$filename

cat > /etc/$filename/config.yaml <<EOF
port: 9126
timeout_server: 5
metrics:
  - process_name: "cpu_usage"
    command: "grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage }' |  xargs printf '%.*f\\n' '2'"
  - process_name: "ram_usage"
    command: "free | grep Mem | awk '{print $3/$2 * 100}' | xargs printf '%.*f\\n' '2'"
  - process_name: "disk_usage"
    command: "df -h --total | tail -n 1 | awk '{print $5}' | tr -d '%'"
  - process_name: "cpu_temp"
    command: "vcgencmd measure_temp | sed -E 's/[^0-9]+C//' | sed -E 's/[^0-9]+//'"
EOF

chown $filename:$filename /etc/$filename/config.yaml
chmod 750 -R /etc/$filename

cat > /etc/systemd/system/$filename.service <<EOF
[Unit]
Description=$filename
Wants=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/$filename/config.yaml

[Service]
User=$filename
Group=$filename
Type=simple
ExecStart=/usr/local/bin/$filename -config.file=/etc/$filename/config.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

chmod 640 /etc/systemd/system/$filename.service

systemctl daemon-reload
systemctl start $filename
systemctl enable $filename