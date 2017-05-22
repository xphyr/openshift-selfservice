import {ModuleWithProviders} from '@angular/core';
import {Routes, RouterModule} from '@angular/router';
import {HomeComponent} from './home/home.component';
import {LoginComponent} from './login/login.component';
import {CanActivateViaAuthGuard} from './core/auth/auth.guard';
import {EditQuotasComponent} from './openshift/edit-quotas/edit-quotas.component';

export const appRoutes: Routes = [
  {path: 'home', component: HomeComponent, canActivate: [CanActivateViaAuthGuard]},
  {path: 'openshift/editquotas', component: EditQuotasComponent, canActivate: [CanActivateViaAuthGuard]},
  {path: 'login', component: LoginComponent},
  {path: '**', component: LoginComponent}
];

export const routing: ModuleWithProviders = RouterModule.forRoot(appRoutes, {useHash: true});
