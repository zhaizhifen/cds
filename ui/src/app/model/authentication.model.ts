import { User } from './user.model';

export class AuthDriverManifest {
    type: string;
    signup_disabled: boolean;
}

export class AuthConsumerSigninResponse {
    token: string;
    user: User;
}
export class AuthConsumer {
    id: string;
    name: string;
    description: string;
    parent_id: string;
    authentified_user_id: string;
    type: string;
    created: string;
    group_ids: Array<number>;
    scopes: Array<string>;
}