VOKAL Image Proxy
===========

`vip` is an image proxy designed for easily resizing and caching images.

Images are served up with a URI that contains an S3 bucket name, as well as a 
unique identifier for the image:
        
        www.example.com/mybucket/5272a0e7d0d9813e21000009-1386617305123315142

When images are requested they are placed into an in-memory cache to make repeated
requests for that image faster.

You can resize an image on the fly by providing an `?s=X` parameter that specifies
a maximum width for the image. The maximum width that can be provided is 720 pixels.
For example, if you want to resize an image down to a 160 pixel thumbnail:
        
        www.example.com/mybucket/5272a0e7d0d9813e21000009-1386617305123315142?s=160

The thumbnail will then be cached to both `groupcache` and S3. If the image leaves
the in-memory cache it will not need to be resized again. 

### Deploying

`vip` can be deployed with Docker:

        sudo docker run -e DATABASE_URL=... -e AWS_SECRET_ACCESS_KEY=... -e AWS_ACCESS_KEY_ID=... vokalinteractive/vip 
