# Weave

Weave is a tool that allows you to generate configuration packages for your varying configuration groups. It merges your base with each configuration group, while also applying exclusions.

Weave produces Tar Gzip archives (optional AES encrypted) which can be extracted onto your target for instant use.

Weave is useful if you have a core configuration repository and want to distribute sensitive parts to specific hosts.

I use Weave to distribute sensitive Dockerfiles to various docker hosts, ensuring they only get the Dockerfiles they need.

## Weave layout

Your weave project should look like this

* config.yml (the configuration file)
* configurations (the confiuration groups you *may* want to merge in)
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
```

This configuration will produce two packages docker1.callumj.com and docker2.callumj.com. docker1.callumj.com will include everything except anything starting with extend-data, with docker2.callumj.com including only files starting with extended-data.

We don't want to include any git metadata, OS X junk or bundler stuff so we've excluded that.

Because I'm dogfooding I've also enabled encryption (you're free to disable this and employ your own encryption after the fact)

My weave setup also includes this

* configurations/docker1.callumj.com/radio1/id_rsa.pub
* configurations/docker1.callumj.com/radio1/id_rsa
* configurations/docker2.callumj.com/extended-data/s3_conf.yml

For which these will be merged into their respestive packages.

I then call weave from this working directory

```
weave config.yml
```

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
