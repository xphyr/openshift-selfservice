import {BrowserModule} from '@angular/platform-browser';
import {NgModule} from '@angular/core';

import {AppComponent} from './app.component';
import {CoreModule} from "./core/core.module";
import {SimpleNotificationsModule} from "angular2-notifications";
import {LoginComponent} from './login/login.component';
import {routing} from './app.routes';
import { HomeComponent } from './home/home.component';
import { EditQuotasComponent } from './openshift/edit-quotas/edit-quotas.component';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    HomeComponent,
    EditQuotasComponent
  ],
  imports: [
    BrowserModule,
    CoreModule,
    SimpleNotificationsModule.forRoot(),
    routing
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule {
}
