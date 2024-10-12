3. Install Chrome and ChromeDriver on Linux
On your Linux machine, run the following commands to install Chrome and ChromeDriver:

sudo apt update
sudo apt install -y wget unzip
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo apt install -y ./google-chrome-stable_current_amd64.deb
rm google-chrome-stable_current_amd64.deb

wget https://chromedriver.storage.googleapis.com/114.0.5735.90/chromedriver_linux64.zip
unzip chromedriver_linux64.zip
sudo mv chromedriver /usr/local/bin/
rm chromedriver_linux64.zip

 Set Up Execution Permissions
chmod +x bin-collector-symbio-linux

. Run the Application
./bin-collector-symbio-linux