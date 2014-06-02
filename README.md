![weave](http://metrix.callumj.com/metric/increment?key=image_redirect&source=weave&redirect=http://weavefiles.callumj.com/static/logo1.png)

Weave is a tool that allows you to generate configuration packages for your varying configuration groups. It merges your base with each configuration group, while also applying exclusions.

* References a core configuration repository and want to distribute sensitive parts to specific hosts.
* Tar Gzip archives (optional AES encrypted) which can be extracted onto your target for instant use.
* Supports downloading packages from HTTP endpoints, if the package is stored with an eTag Weave will only download the package after a change is made. Inversely you can use Weave's built in S3 support to automatically upload modified files after generation.
* Support for `post_extraction.sh` and `pre_extraction.sh` callback scripts, which will be called before and after extraction **when** a new configuration is downloaded.

I use Weave to generate configuration groups for specific Docker hosts, ensuring they only get the sensitive Dockerfiles they need and nothing more.

## Installation

Binaries are available from [weavefiles.callumj.com](http://weavefiles.callumj.com/)

Alternatively you can use Go to install `weave`

```
go install github.com/callumj/weave
```

## Weave layout

Your weave project should look like this

* config.yml (the configuration file)
* configurations (the configuration groups you *may* want to merge in)
* working (the output directory)
* keys (optional folder for encryption keys)

### config.yml

```yaml
src: /Users/callumj/Development/DockerConf
encrypt: true
configurations:
  - name: docker1.callumj.com
    except:
      - ^extended-data
  - name: docker2.callumj.com
    only:
      - ^extended-data
ignore:
  - ^\.git
  - /\.git
  - \.DS_Store
  - ^\.bundle
  - /\.bundle
s3:
  access_key: AMZASSD456822X2
  secret: "ShussThis5sASe3crt"
  bucket: myweavebucket
  endpoint: eu-west-1
  folder: docker
  public: true
```

This configuration will produce two packages docker1.callumj.com and docker2.callumj.com. docker1.callumj.com will include everything except anything starting with extend-data, with docker2.callumj.com including only files starting with extended-data.

We don't want to include any git metadata, OS X junk or bundler stuff so we've excluded that.

Because I'm dogfooding I've also enabled encryption (you're free to disable this and employ your own encryption after the fact)

My weave setup also includes this

* configurations/docker1.callumj.com/radio1/id_rsa.pub
* configurations/docker1.callumj.com/radio1/id_rsa
* configurations/docker2.callumj.com/extended-data/s3_conf.yml

For which these will be merged into their respestive packages. Weave doesn't require anything to merge with, which is useful if you only want to filter configuration packages.

Updated packages will be uploaded to the S3 bucket when their hashed filenames change.

I then call weave from this working directory

```
weave config.yml
```

You can also pass the following command line arguments in tweak how weave performs.

* `-n` Disable pushing to S3
* `-o CONFIGURATION_NAME` Generate only the specified configuration

## Workflow

Weave runs in this order

* Configuration parsed
* A base tar (no gzip) is created of the source, versioned into `working/`
* For each configuration group the base tar is loaded in, exlcusions/filters are applied and then for the files that exist in `configurations/CONFIGURATION_NAME/` these files are added to the archive
* The archive is gzipped and the optionally encrypted, outputting into `working/`

## Extracting

If your archive is encrypted then you will need to use Weave to extract your configuration.

```
weave /path/to/package2222.tar.gz.enc /path/to/keyfile.txt /out/put/directory
```

Alternatively if I wanted to make use of S3, I could pass in a standard HTTP address (it doesn't have to be from S3)

```
weave http://s3-eu-west-1.amazonaws.com//Datacom/Kimberley/package2222.tar.gz.enc /path/to/keyfile.txt /out/put/directory
```

### Callbacks

When a new package is available (the only case where this is not true is when ETags match, causing no new downloads) callback scripts are invoked pre and post extraction.

These callbacks are called with directory set to the final extraction directory and run using `/bin/bash`.

* `pre_extraction.sh`: Called just before the extraction begins, you should use to this clean any files that may not exist in the next extraction.
* `post_extraction.sh`: Called after extraction, allowing you to invoke any scripts you need to perform such as deployment tools (Ansible, Chef, etc) or launch the application.

STDOUT and STDERR are captured from the callbacks and are stored as *.log files in the output directory.
