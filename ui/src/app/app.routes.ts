import {ModuleWithProviders} from '@angular/core';
import {Routes, RouterModule} from '@angular/router';
import {HomeComponent} from "./home/home.component";
import {LoginComponent} from "./login/login.component";
import {CanActivateViaAuthGuard} from "./core/auth/auth.guard";

export const appRoutes: Routes = [
  {path: 'login', component: LoginComponent},
  {path: '**', component: HomeComponent, canActivate: [CanActivateViaAuthGuard]}
];

export const routing: ModuleWithProviders = RouterModule.forRoot(appRoutes, {useHash: true});
