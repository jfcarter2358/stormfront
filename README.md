```
__     __   _            _ __  __           _      _
\ \   / /__| | ___   ___(_)  \/  | ___   __| | ___| |
 \ \ / / _ \ |/ _ \ / __| | |\/| |/ _ \ / _` |/ _ \ |
  \ V /  __/ | (_) | (__| | |  | | (_) | (_| |  __/ |
   \_/ \___|_|\___/ \___|_|_|  |_|\___/ \__,_|\___|_|
```

# About

Velocimodel is a comprehensive open-source model operations solution. It allows for the management and versioning of data science models (in addition to other software projects) and will eventually allow for model automation and deployment.

# Running Locally

To run VelociModel locally, clone this repo and make sure you have Python3 and docker-compose installed. After pulling down the repo, run

```
pip install -r requirements.txt
```

then

```
make build-docker
```

To build local versions of the various docker images. Finally, run

```
make run-docker
```

to bring up the services and navigate to `http://localhost:9000` to view the UI.

# To Do

**v0.1.0**

- [x] service-manager
   - [x] service registry
   - [x] config storage
   - [x] secret storage
- [x] asset-manager
   - [x] asset registry
   - [x] local file upload
   - [x] local file download
   - [x] automatic git sync
   - [x] manual git sync
   - [x] git asset creation
   - [x] pull config from service-manager
- [x] model-manager
   - [x] model registry
   - [x] model snapshots
   - [x] model releases
   - [x] pull config from service-manager
   - [x] model download
   - [x] snapshot download
   - [x] release download
- [x] stormfront
- [x] frontend
   - [x] dashboard
   - [x] model view
   - [x] model edit
   - [x] model creation
   - [x] model add existing asset
   - [x] model add new file asset
   - [x] model add new git asset
   - [x] model delete asset
   - [x] model download
   - [x] model code save
   - [x] snapshot view
   - [x] snapshot edit
   - [x] snapshot creation
   - [x] snapshot download
   - [x] snapshot code save
   - [x] release view
   - [x] release creation
   - [x] release download
   - [x] asset view
   - [x] asset edit
   - [x] asset download
   - [x] asset code save
   - [x] file asset creation
   - [x] git asset creation

**v0.2.0**

- [x] frontend
   - [x] Param view/edit
   - [x] Secret view/edit
   - [x] User view
   - [x] User edit view
   - [x] User create view
   - [x] User delete view
   - [x] Service status badge
   - [x] BasePath support
- [x] auth-manager
   - [x] Authorization flow
   - [x] BasePath support
- [x] stormfront
   - [x] All endpoints secured through stormfront
- [x] asset-manager
   - [x] SSH git clone support
- [x] service-manager
   - [x] Automated service health pings

**v0.3.0**

- [ ] whitetail integration
- [ ] velocimodel-sdk
- [ ] frontend
   - [ ] Better error propagation
   - [ ] Add secrets view
   - [ ] Delete secrets view
   - [ ] Add params view
   - [ ] Delete params view
- [ ] automation-manager
- [ ] runtime-manager
- [ ] model-runtime
- [ ] auth-manager
   - [ ] password reset
   - [ ] group-based access
   - [ ] role-based access
- [ ] asset-manager
   - [ ] Convert file download to a redirect
- [ ] service-manager
   - [ ] Convert to using IPs instead of hostnames

**v0.4.0**

- [ ] asset-manager
   - [ ] S3 integration
   - [ ] Ability to mark assets as sensitive
      - Disables download of that asset
- [ ] auth-manager
   - [ ] API token generation and authentication

**v0.5.0**

- [ ] asset-manager
   - [ ] Artifactory integration

**v0.6.0**

- [ ] asset-manager
   - [ ] Azure Blob Store integration

**v1.0.0**

- Initial Release!
- [ ] Usage documentation
- [ ] API documentation

# Contact

This software is written by John Carter. If you have any questions or concerns feel free to create an issue on GitHub or send me an email at jfcarter2358(at)gmail.com
