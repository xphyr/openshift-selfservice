import {Injectable} from '@angular/core';
import {Http, Response} from "@angular/http";
import {NotificationsService} from "angular2-notifications";
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import {Observable} from "rxjs/Observable";
import 'rxjs/add/observable/throw'
import {Router} from "@angular/router";

@Injectable()
export class AuthService {
  public userData;

  constructor(private http: Http, private notificationService: NotificationsService,
              private router: Router) {

    let existingUserData = localStorage.getItem("AUTH");
    if (existingUserData) {
      this.userData = JSON.parse(existingUserData);
    }
  }

  login(username: string, password: string) {
    this.http.post('/login', {
      username,
      password
    })
      .map((res: Response) => res.json())
      .catch(this.handleError.bind(this))
      .subscribe(data => {
        this.handleToken(data)
      });
  }

  private handleToken(data) {
    // Decode JWT
    let base64Url = data.token.split('.')[1];
    let base64 = base64Url.replace('-', '+').replace('_', '/');
    this.userData = JSON.parse(window.atob(base64));
    console.log('setting', JSON.stringify(this.userData))
    localStorage.setItem('AUTH', JSON.stringify(this.userData));

    this.notificationService.success("Login erfolgreich")
    this.router.navigate(['/home']);
  }

  private handleError(error: Response | any) {
    let msg: string;
    if (error instanceof Response) {
      msg = error.json().message;
    } else {
      msg = error.message ? error.message : error.toString();
    }
    this.notificationService.error("Fehler beim Login", msg);
    return Observable.throw(msg);
  }
}
