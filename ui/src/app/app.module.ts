import {BrowserModule} from '@angular/platform-browser';
import {NgModule} from '@angular/core';

import {AppComponent} from './app.component';
import {CoreModule} from './core/core.module';
import {SimpleNotificationsModule} from 'angular2-notifications';
import {LoginComponent} from './login/login.component';
import {routing} from './app.routes';
import {HomeComponent} from './home/home.component';
import {OpenshiftModule} from './openshift/openshift.module';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    HomeComponent
  ],
  imports: [
    BrowserModule,
    CoreModule,
    SimpleNotificationsModule.forRoot(),
    routing,
    OpenshiftModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule {
}
