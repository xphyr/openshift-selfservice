const quotas = require('./controllers/quotas');

module.exports = function (app, passport) {
    let isAuthenticated = function (req, res, next) {
        if (req.isAuthenticated()) {
            return next();
        }
        res.redirect('/login');
    }

    app.get('/', isAuthenticated, (req, res) => {
        res.render('index.ejs', {message: req.flash('message')});
    });

    app.get('/login', (req, res) => {
        res.render('login.ejs', {message: req.flash('error')});
    });

    app.get('/logout', (req, res) => {
        req.logout();
        res.redirect('/login');
    });

    app.get('/quotas', isAuthenticated, (req, res) => {
        res.render('quotas.ejs');
    });

    app.post('/quotas', isAuthenticated, quotas.updateQuota);

    app.post('/login', passport.authenticate('ldapauth', {
        session: true, successRedirect: '/', failureRedirect: '/login', failureFlash: true
    }));
};
