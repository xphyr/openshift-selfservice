let express = require('express'), passport = require('passport'), bodyParser = require(
'body-parser'), LdapStrategy = require('passport-ldapauth'), cookieParser = require(
'cookie-parser'), session = require('express-session'), flash = require('connect-flash');

// LDAP Options
let OPTS = {
    server: {
        url: process.env.LDAP_URL,
        bindDn: process.env.LDAP_BIND_DN,
        bindCredentials: process.env.LDAP_BIND_CRED,
        searchBase: process.env.LDAP_SEARCH_BASE,
        searchFilter: process.env.LDAP_FILTER
    }
};

let app = express();
app.use(cookieParser());
app.use(session({secret: process.env.SESSION_SECRET}));

passport.use(new LdapStrategy(OPTS));
app.use(passport.initialize());
app.use(passport.session());

// Keeping Sessions in memory
let sessions = new Map();
passport.serializeUser(function(user, done) {
    user.id = user.dn;
    sessions[user.id] = user;
    return done(null, user.id); //this is the 'user' property saved in req.session.passport.user
});

passport.deserializeUser(function (id, done) {
    return done(null, sessions[id]);
});

app.use(flash());

app.use(bodyParser.json());
app.use(bodyParser.urlencoded({extended: false}));
app.set('view engine', 'ejs');
app.set('views', __dirname + '/app/views');

require('./app/routes.js')(app, passport);

app.listen(8080);
console.log('Server running on port 8080');