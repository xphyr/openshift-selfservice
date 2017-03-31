let co = require('co');
let openshift = require('./utils/openshift');

exports.updateQuota = (req, res) => {
    co(function*() {
        yield openshift.checkPermissions(req.user.cn, req.body.project);
        yield openshift.updateProjectQuota(req.user.cn, req.body.project, parseInt(req.body.cpu), parseInt(req.body.memory));

        res.render('quotas.ejs', {
            messages: 'Quota wurde erfolgreich angepasst'
        });
    })
    .catch(err => handleError(err, 'quotas.ejs', res));
};

exports.newProject = (req, res) => {
    co (function*() {
        yield openshift.newProject(req.user.cn, req.body.project, req.body.megaid, req.body.billing);

        res.render('newproject.ejs', {
            messages: 'Projekt wurde erfolgreich angelegt'
        });
    })
    .catch(err => handleError(err, 'newproject.ejs', res));
};

exports.newServiceAccount = (req, res) => {
    co (function*() {
        yield openshift.checkPermissions(req.user.cn, req.body.project);
        yield openshift.newServiceAccount(req.user.cn, req.body.project, req.body.serviceaccount);

        res.render('newserviceaccount.ejs', {
            messages: 'Service-Account wurde erfolgreich angelegt'
        });
    })
    .catch(err => handleError(err, 'newserviceaccount.ejs', res));
};

exports.updateBilling = (req, res) => {
    co (function*() {
        yield openshift.checkPermissions(req.user.cn, req.body.project);
        yield openshift.updateBilling(req.user.cn, req.body.project, req.body.billing);

        res.render('updatebilling.ejs', {
            messages: 'Kontierungsnummer wurde erfolgreich angepasst'
        });
    })
    .catch(err => handleError(err, 'updatebilling.ejs', res));
};

handleError = function(err, page, res){
    if (typeof err.message == 'string') {
        res.render(page, {
            errors: err.message
        });
    } else {
        res.render(page, {
            errors: 'Es ist ein Fehler aufgetreten.'
        });
    }
};

