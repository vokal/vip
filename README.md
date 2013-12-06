Cyfr-Image-Proxy
================

### Environment Setup

Cyfr-Image-Proxy is a Go application thatis designed to run on any Linux-based VPS.

The following steps cover all the required setup for running this project locally.

#### Step One - Build and install

    $ go get github.com/vokalinteractive.com/Cyfr-Image-Proxy

#### Step Two - Local Database

For local installations, Cyfr-Image-Proxy will look at `localhost` for it's MongoDB
 instance. There is no configuration necessary.

### Running the Local Web Server

    $ AWS_ACCESS_KEY=myaccesskey AWS_SECRET_KEY=mysecretkey Cyfr-Image-Proxy

### Sample Usage

Assuming that the server is hosted at `images.cyfr.net`, when you upload an image to
the Cyfr service you will be given a serving URL for where to fetch that image. The
serving URL proxies the image through the Cyfr-Image-Proxy. The following example URL:

    http://images.cyfr.net/cyfr-chat/521411e8169cb86491000001

will serve up the original image, as it was uploaded. If a requesting client wants to
display a smaller version of this image (ex: for a thumbnail) it can be resized on the
fly by appending a `s=XX` querystring parameter, where `XX` is the width you would like.
Images are scaled with their aspect ratio in-tact:

    http://images.cyfr.net/cyfr-chat/521411e8169cb86491000001?s=250

Once an image has been proxied through the service it will be cached. Initial resizes
will always be slow. Images will be cached for 1 week.
