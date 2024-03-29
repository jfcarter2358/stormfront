```
 ____  _                        __                 _   
/ ___|| |_ ___  _ __ _ __ ___  / _|_ __ ___  _ __ | |_ 
\___ \| __/ _ \| '__| '_ ` _ \| |_| '__/ _ \| '_ \| __|
 ___) | || (_) | |  | | | | | |  _| | | (_) | | | | |_ 
|____/ \__\___/|_|  |_| |_| |_|_| |_|  \___/|_| |_|\__|
```

# About

Stormfront is a container orchestration system designed around 

# Install

```bash
# Requires JQ installed on your machine
curl -s https://raw.githubusercontent.com/jfcarter2358/stormfront/main/install.sh > install.sh
chmod +x install.sh
sudo ./install.sh
rm install.sh
```

# Uninstall

```bash
curl -s https://raw.githubusercontent.com/jfcarter2358/stormfront/main/uninstall.sh > uninstall.sh
chmod +x uninstall.sh
sudo ./uninstall.sh
rm uninstall.sh
```

# TODO

## 1.0.0

**Features**

- [x] Check system health
- [x] Create application
- [x] Restart application
- [x] Delete application
- [x] Send command to node
- [x] Generate join command
- [x] Check system resource usage
- [x] Application persistence
- [ ] Persistence replication
- [x] Service DNS
- [ ] Image pull policies
- [ ] Image restart policies
- [ ] Disaster recovery
- [ ] Docker credentials
- [ ] Secrets management
- [ ] Log trailing

**Bugs**

- None

# Contact

This software is written by John Carter. If you have any questions or concerns feel free to create an issue on GitHub or send me an email at jfcarter2358(at)gmail.com
