import {Component, OnInit} from '@angular/core';
import {OpenshiftService} from '../openshift.service';
import {NotificationsService} from 'angular2-notifications';

@Component({
  selector: 'app-edit-quotas',
  templateUrl: './edit-quotas.component.html'
})
export class EditQuotasComponent implements OnInit {
  public projectName: string;
  public cpu: number;
  public memory: number;

  constructor(private openshiftService: OpenshiftService, private notificationsService: NotificationsService) {
  }

  ngOnInit() {
  }

  updateQuota() {
    this.openshiftService.updateQuotas(this.projectName, this.cpu, this.memory)
      .catch(err => {
          this.notificationsService.alert('Fehler beim Update', err.json().message);
      })
      .subscribe(r => {
        console.log('r', r);
      });
  }
}
