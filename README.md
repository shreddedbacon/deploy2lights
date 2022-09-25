# lagoondeploy2lights

docker build -t uselagoon/deploy2lights .


## setup raspberry pi os (lite)

Configure WIFI and SSH using raspberry pi imager (https://www.raspberrypi.com/software/) install raspberry pi os lite 64bit

```
sudo apt-get install git
wget https://go.dev/dl/go1.18.6.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.18.6.linux-arm64.tar.gz
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker pi

sudo curl -SL https://github.com/docker/compose/releases/download/v2.11.1/docker-compose-linux-armv7 -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```
