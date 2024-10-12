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

chmod +x ./chromedriver-linux64/chromedriver

o run bin-collector in Linux, follow these steps:

Build the Go application: If you haven't already built the Go app, navigate to your project folder and run:

bash
Copy code
go build -o bin-collector-symbio-linux main.go
Make the chromedriver executable: Ensure that chromedriver has execution permission by running:

bash
Copy code
chmod +x ./chromedriver-linux64/chromedriver
Run the application: You can then run the application using the following command:

bash
Copy code
./bin-collector-symbio-linux
The web server should now be running at http://0.0.0.0:8080. Make sure chromedriver and Google Chrome are correctly installed in your system.

If you're using Docker, build and run the container as follows:

Build the Docker image:

bash
Copy code
docker build -t bin-collector .
Run the Docker container:

bash
Copy code
docker run -p 8080:8080 bin-collector
This will start your app inside the container and expose port 8080 for access across your LAN or localhost.

