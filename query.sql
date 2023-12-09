

DELETE FROM "user_project" 
WHERE project_id = '1acd7a8f-a038-4e07-91cb-b689c368d855'
    AND company_id = '368b73a9-beb3-4c92-ba9f-3188ee387103'
    AND user_id = '7e33f2f0-cb34-4efa-b930-3bb35bce39db'
    AND client_type_id = '52e53be2-9959-414b-b720-ac6c10e5ad3e'
    AND role_id = '7e33f2f0-cb34-4efa-b930-3bb35bce39db'



SELECT
    project_id,
    env_id
from user_project
where user_id = '923752e7-10a4-442b-82a3-614031bd7f49'
