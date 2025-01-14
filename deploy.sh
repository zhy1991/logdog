#!/bin/bash
deploy_name=Logdog
if [ ! -d "/opt/${deploy_name}" ]; then
        echo "应用目录 ${deploy_name} 不存在 ,自动创建 ${deploy_name}"
        mkdir -p /opt/${deploy_name}
fi
cp config_demo.yaml /opt/${deploy_name}/
cp ${deploy_name}  /opt/${deploy_name}/${deploy_name}
 
cat > /usr/lib/systemd/system/${deploy_name}.service << EOF
[Unit]
Description=${deploy_name}

[Service]
WorkingDirectory=/opt/${deploy_name}/
PrivateTmp=true
Restart=always
Type=simple
ExecStart=/opt/${deploy_name}/${deploy_name} -c /opt/${deploy_name}/config.yaml
ExecStop=/usr/bin/kill -15 $MAINPID
User=root
Group=root
LimitCORE=infinity
LimitNOFILE=65535
LimitNPROC=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl  daemon-reload
