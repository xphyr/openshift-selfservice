let rp = require('request-promise');

let OSE_API = process.env.OPENSHIFT_API_URL;
let OSE_TOKEN = process.env.OPENSHIFT_TOKEN;

let MAX_CPU = process.env.MAX_CPU;
let MAX_MEMORY = process.env.MAX_MEMORY;

let SSPError = require('./ssp-error');

// Add helpful things :)
String.prototype.isEmpty = function () {
    return (this.length === 0 || !this.trim());
};

exports.getHttpOpts = function (uri) {
    return {
        uri: uri, rejectUnauthorized: false, headers: {
            'Authorization': 'Bearer ' + OSE_TOKEN
        }, json: true
    };
};

exports.checkPermissions = function (username, project) {
    return rp(this.getHttpOpts(`${OSE_API}/oapi/v1/namespaces/${project}/policybindings/:default/`))
    .then(res => {
        // Check if a User is admin
        let isAdmin = false;
        if (res && res.roleBindings) {
            res.roleBindings.forEach(rb => {
                if (rb.name === 'admin') {
                    rb.roleBinding.userNames.forEach(un => {
                        if (un.toLowerCase() === username.toLowerCase()) {
                            isAdmin = true;
                        }
                    })
                }
            })
        }

        if (!isAdmin) {
            console.error(`User ${username} cannot edit project ${project} as he has no admin rights`);
            return Promise.reject('Du hast auf dem Projekt keine Admin-Rechte');
        }
    })
    .catch((err) => {
        if (typeof err === 'string') {
            throw new SSPError(err);
        }
        console.error(err.message);
        throw new SSPError('Projekt konnte nicht gefunden werden');
    });
};

exports.updateProjectQuota = function (username, project, cpu, memory) {
    if (project.isEmpty()) {
        throw new SSPError('Projektname muss angegeben werden');
    }

    if (cpu > MAX_CPU) {
        throw new SSPError(`Es können maximal ${MAX_CPU} CPU Cores vergeben werden.`);
    }

    if (memory > MAX_MEMORY) {
        throw new SSPError(`Es können maximal ${MAX_MEMORY}GB Memory vergeben werden.`);
    }

    // Get existing quota
    return rp(this.getHttpOpts(`${OSE_API}/api/v1/namespaces/${project}/resourcequotas`)).then(res => {
        let quota = res.items[0];
        quota.spec.hard.cpu = cpu;
        quota.spec.hard.memory = memory + 'Gi';

        // Update quota
        let httpOpts = this.getHttpOpts(`${OSE_API}/api/v1/namespaces/${project}/resourcequotas/${quota.metadata.name}`);
        httpOpts.method = 'PUT';
        httpOpts.body = quota;

        return rp(httpOpts).then(() => {
            console.log(`User ${username} changed quotas for project ${project}. CPU: ${cpu}, Mem: ${memory}`);
            return Promise.resolve();
        }, (err) => {
            console.error(err);
            return Promise.reject(err);
        });
    });
};

exports.newProject = function (username, project, isTestproject, megaId, billing) {
    if (!isTestproject && project.isEmpty()) {
        throw new SSPError('Projektname muss angegeben werden');
    }

    if (!isTestproject && billing.isEmpty()) {
        throw new SSPError('Kontierungsnummer muss angegeben werden');
    }

    // Always format project name like this: username-project
    if (isTestproject) {
        project = `${username}-${project}`;
        billing = 'keine-verrechnung';
    }

    let httpOpts = this.getHttpOpts(`${OSE_API}/oapi/v1/projectrequests`);
    httpOpts.method = 'POST';
    httpOpts.body = {
        'kind': 'ProjectRequest', 'apiVersion': 'v1', 'metadata': {
            'name': project
        }
    };
    return rp(httpOpts).then(() => {
        console.log(`User ${username} created a new project: ${project}`);
        return Promise.resolve();
    })
    .then(() => this.changePermissions(project, username))
    .then(() => this.updateMetadata(project, billing, megaId, username))
    .catch(err => {
        console.error('Error occured while creating project: ', err.message);
        if (err.statusCode === 409) {
            return Promise.reject(new SSPError('Das Projekt existiert bereits'));
        }

        return Promise.reject(err);
    })
};

exports.newServiceAccount = function (username, project, serviceAccount) {
    if (project.isEmpty()) {
        throw new SSPError('Projektname muss angegeben werden');
    }

    if (serviceAccount.isEmpty()) {
        throw new SSPError('Service-Account-Name muss angegeben werden');
    }

    let httpOptions = this.getHttpOpts(`${OSE_API}/api/v1/namespaces/${project}/serviceaccounts`);
    httpOptions.method = 'POST';
    httpOptions.body = {
        "kind": "ServiceAccount", "apiVersion": "v1", "metadata": {
            "name": serviceAccount
        }
    };
    return rp(httpOptions).then(() => {
        console.log(`User ${username} has created a serviceAccount for project ${project}. Name: ${serviceAccount}`);
        return Promise.resolve();
    })
    .catch(err => {
        console.log('Error while creating serviceAccount for project: ', err.message);
        return Promise.reject(err);
    });
};

exports.updateBilling = function (username, project, billing) {
    if (project.isEmpty()) {
        throw new SSPError('Projektname muss angegeben werden');
    }

    if (billing.isEmpty()) {
        throw new SSPError('Kontierungsnummer muss angegeben werden');
    }

    return this.updateMetadata(project, billing, null, username);
};

exports.updateMetadata = function (project, billing, megaId, username) {
    let url = `${OSE_API}/api/v1/namespaces/${project}`;
    return rp(this.getHttpOpts(url))
    .then(res => {

        // Update metadata
        res.metadata.annotations['openshift.io/kontierung-element'] = billing;
        if (megaId) {
            res.metadata.annotations['openshift.io/MEGAID'] = megaId;
        }

        let httpOpts = this.getHttpOpts(url);
        httpOpts.method = 'PUT';
        httpOpts.body = res;
        return rp(httpOpts).then(() => {
            console.log(`User ${username} changed metadata of project ${project}. Billing: ${billing}, MEGA-ID: ${megaId}`);
            return Promise.resolve();
        })
        .catch(err => {
            console.log('Error while updating project metadata: ', err.message);
            return Promise.reject(err);
        });
    })
};

exports.changePermissions = function (project, username) {
    let url = `${OSE_API}/oapi/v1/namespaces/${project}/policybindings/:default/`;
    return rp(this.getHttpOpts(url))
    .then(res => {
        debugger;

        if (res && res.roleBindings) {
            res.roleBindings.forEach(rb => {
                if (rb.name === 'admin') {
                    rb.roleBinding.userNames[0] = username;
                    rb.roleBinding.subjects[0] = {
                        'kind': 'User', 'name': username
                    }
                }
            })
        }

        let httpOpts = this.getHttpOpts(url);
        httpOpts.method = 'PUT';
        httpOpts.body = res;
        return rp(httpOpts).then(() => {
            console.log(`User ${username} is now admin of project ${project}`);
            return Promise.resolve();
        })
        .catch(err => {
            console.log('Error while chaning admin permissions: ', err.message);
            return Promise.reject(err);
        });
    });
};
