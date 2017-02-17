let rp = require('request-promise');

let OSE_API = process.env.OPENSHIFT_API_URL;
let OSE_TOKEN = process.env.OPENSHIFT_TOKEN;

let MAX_CPU = process.env.MAX_CPU;
let MAX_MEMORY = process.env.MAX_MEMORY;

exports.getHttpOpts = function (uri) {
    return {
        uri: uri, rejectUnauthorized: false, headers: {
            'Authorization': 'Bearer ' + OSE_TOKEN
        }, json: true
    };
}

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
            throw new Error(err);
        }
        throw new Error('Projekt konnte nicht gefunden werden');
    });
}

exports.updateProjectQuota = function (username, project, cpu, memory) {
    if (project.length === 0) {
        throw new Error('Projektname muss angegeben werden');
    }

    if (cpu > MAX_CPU) {
        throw new Error(`Es können maximal ${MAX_CPU} CPU Cores vergeben werden.`);
    }

    if (memory > MAX_MEMORY) {
        throw new Error(`Es können maximal ${MAX_MEMORY}GB Memory vergeben werden.`);
    }

    // Get existing quota
    return rp(this.getHttpOpts(`${OSE_API}/api/v1/namespaces/${project}/resourcequotas`)).then(res => {
        let quota = res.items[0];
        quota.spec.hard.cpu = cpu;
        quota.spec.hard.memory = memory + 'Gi';

        // Update quota
        let httpOpts = this.getHttpOpts(
        `${OSE_API}/api/v1/namespaces/${project}/resourcequotas/${quota.metadata.name}`);
        httpOpts.method = 'PUT';
        httpOpts.body = quota;

        return rp(httpOpts).then(() => {
            console.log(
            `User ${username} changed quotas for project ${project}. CPU: ${cpu}, Mem: ${memory}`);
            return Promise.resolve();
        }, (err) => {
            console.error(err);
            return Promise.reject(err);
        });
    });
};
