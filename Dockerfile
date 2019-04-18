FROM golang:1.12 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV TZ=Asia/Tokyo

COPY . /
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add \
 && echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list \
 && apt-get update && apt-get -y install google-chrome-stable unzip \
 && CHROME_DRIVER_VERSION=`curl -sS chromedriver.storage.googleapis.com/LATEST_RELEASE` \
 && wget -N http://chromedriver.storage.googleapis.com/$CHROME_DRIVER_VERSION/chromedriver_linux64.zip -P ~/ \
 && unzip ~/chromedriver_linux64.zip -d ~/ \
 && rm ~/chromedriver_linux64.zip \
 && chown root:root ~/chromedriver \
 && chmod 755 ~/chromedriver \
 && mv ~/chromedriver /usr/bin/chromedriver \
 && apt-get -y remove unzip \
 && apt clean && rm -rf /var/lib/apt/lists/* \
 && go build -o /main /main.go
CMD /main $PORT

