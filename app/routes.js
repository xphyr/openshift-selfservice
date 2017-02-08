module.exports = function (app, passport) {
    let isAuthenticated = function (req, res, next) {
        if (req.isAuthenticated()) {
            return next();
        }
        res.redirect('/');
    }

    app.get('/', (req, res) => {
        res.render('index.ejs', {message: req.flash('message')});
    });

    app.get('/login', (req, res) => {
        res.render('login.ejs', {message: req.flash('error')});
    });

    app.get('/quotas', isAuthenticated, (req, res) => {
        res.render('quotas.ejs');
    });

    app.get('/logout', (req, res) => {
        req.logout();
        res.redirect('/');
    });

    app.post('/login', passport.authenticate('ldapauth', {
        session: true, successRedirect: '/quotas', failureRedirect: '/login', failureFlash: true
    }));
};
