let co = require('co');
let openshift = require('./../utils/openshift');

exports.updateQuota = (req, res) => {
    co(function*() {
        yield openshift.checkPermissions(req.user.cn, req.body.project);
        yield openshift.updateProjectQuota(req.user.cn, req.body.project, parseInt(req.body.cpu), parseInt(req.body.memory));

        res.render('quotas.ejs', {
            messages: 'Quota wurde erfolgreich angepasst'
        });
    })
    .catch(err => {
        res.render('quotas.ejs', {
            errors: err.message
        });
    });
};