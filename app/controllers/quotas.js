exports.updateQuota = (req, res) => {

    console.log(req.body('cpu'));

    // TODO: Go to Openshift API

    res.render('quotas.ejs', {message: 'hi'});
};