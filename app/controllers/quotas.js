let openshift = require('./../utils/openshift');

exports.updateQuota = (req, res) => {
    openshift.checkPermissions("u220374", "ose-mon-a")
             .then(openshift.updateProjectQuota("u220374", "ose-mon-a", req.body.cpu, req.body.memory))
             .then(() => {
                 res.render('quotas.ejs', {
                     messages: 'Quota wurde erfolgreich angepasst'
                 });
             })
             .catch((err) => {
                 res.render('quotas.ejs', {
                     errors: err
                 });
             });
};