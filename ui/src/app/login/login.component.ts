import { Component, OnInit } from '@angular/core';
import {AuthService} from "../core/auth/auth.service";

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styles: [`
     body {
            padding-top: 80px;
        }
  `]
})
export class LoginComponent implements OnInit {

  public username: string;
  public password: string;

  constructor(private authService: AuthService) { }

  ngOnInit() {
  }

  login() {
    this.authService.login(this.username, this.password);
  }

}
