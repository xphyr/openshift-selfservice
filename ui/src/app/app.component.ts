import { Component } from '@angular/core';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html'
})
export class AppComponent {
  public showNavbar: boolean = true;

  public notificationOptions = {
    position: ['top', 'right'],
    timeOut: 5000,
    maxStack: 5,
    lastOnBottom: true
  };
}
